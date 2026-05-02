package usecase

import (
    "context"
    "fmt"
    "strings"
    "time"

    "backend/configs"
    "backend/internals/chatbot/controller/dto"
    "backend/internals/chatbot/tools"
    "backend/pkgs/ai"
    pkgredis "backend/pkgs/redis"

    "github.com/google/uuid"
)

const chatSystemPrompt = `Bạn là trợ lý SQL thông minh tên ChamsBot, hỗ trợ sinh viên học SQL trong hệ thống ChamsQL.

NHIỆM VỤ:
1. Giúp sinh viên hiểu lỗi và cách sửa câu SQL
2. Gợi ý hướng tiếp cận bài toán — KHÔNG cho đáp án trực tiếp
3. Giải thích SQL concept bằng tiếng Việt dễ hiểu có ví dụ
4. Dùng tools để lấy thông tin thực tế trước khi trả lời

NGUYÊN TẮC SỬ DỤNG TOOLS:
- Khi sinh viên hỏi về bài cụ thể → gọi get_problem_schema trước
- Khi sinh viên paste SQL bị lỗi → gọi run_student_sql để xem lỗi thật
- Khi cần so sánh → gọi compare_with_solution, KHÔNG tiết lộ nội dung đáp án
- Khi hỏi về concept → gọi explain_sql_concept

PHONG CÁCH TRẢ LỜI:
- Tiếng Việt, thân thiện, khuyến khích
- Dùng Markdown: code block cho SQL, bullet points cho danh sách
- ⚠️ BẮT BUỘC khi dùng bảng Markdown: Mỗi dòng bảng phải xuống dòng riêng để hiển thị đúng.
  Ví dụ:
  | Cột 1 | Cột 2 |
  |-------|-------|
  | Data  | Data  |
- Ngắn gọn, có ví dụ cụ thể
- Kết thúc bằng câu hỏi gợi ý hoặc bước tiếp theo`

type IChatbotUseCase interface {
    Ask(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error)
}

type chatbotUseCase struct {
    cfg      *configs.Config
    aiClient ai.IChatLLMClient
    executor *tools.ToolExecutor
    redis    pkgredis.IRedis
}

func NewChatbotUseCase(cfg *configs.Config, executor *tools.ToolExecutor, redis pkgredis.IRedis) IChatbotUseCase {
    var aiClient ai.IChatLLMClient
    if cfg.OpenAIAPIKey != "" {
        aiClient = ai.NewOpenAIChatClient(cfg.OpenAIAPIKey, cfg.OpenAIBaseURL, cfg.OpenAIModel)
    }
    return &chatbotUseCase{
        cfg:      cfg,
        aiClient: aiClient,
        executor: executor,
        redis:    redis,
    }
}

func (u *chatbotUseCase) Ask(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error) {
	start := time.Now()

	// Rate limit: tối đa 30 requests/user/giờ
	if req.UserID != nil && u.redis != nil {
		rateLimitKey := fmt.Sprintf("chatbot_rl:%d", *req.UserID)
		var entry struct {
			Count    int       `json:"count"`
			ExpireAt time.Time `json:"expireAt"`
		}
		
		now := time.Now()
		if err := u.redis.Get(rateLimitKey, &entry); err != nil {
			// Lần đầu gọi trong window
			entry.Count = 0
			entry.ExpireAt = now.Add(1 * time.Hour)
		} else if now.After(entry.ExpireAt) {
			// Đã qua 1 giờ, reset
			entry.Count = 0
			entry.ExpireAt = now.Add(1 * time.Hour)
		}
		
		if entry.Count >= 30 {
			return nil, fmt.Errorf("bạn đã dùng hết lượt hỗ trợ AI trong giờ này (tối đa 30 lượt/giờ). Vui lòng thử lại sau")
		}
		
		entry.Count++
		// Tăng count và giữ nguyên TTL cũ bằng time.Until
		ttl := time.Until(entry.ExpireAt)
		if err := u.redis.SetWithExpiration(rateLimitKey, entry, ttl); err != nil {
			fmt.Printf("Failed to update rate limit counter: %v\n", err)
		}
	}

	conversationID := req.ConversationID
    if conversationID == "" {
        conversationID = uuid.New().String()
    }

    // Fallback nếu không có OpenAI key
    if u.aiClient == nil {
        reply := u.patternResponse(req)
        return &dto.ChatResponse{
            Reply:          reply,
            ConversationID: conversationID,
            Provider:       "pattern",
            ResponseTimeMs: time.Since(start).Milliseconds(),
        }, nil
    }

    // Load history từ Redis
    history := u.loadHistory(conversationID)

    // Build user message với context
    userContent := u.buildUserMessage(req)
    history = append(history, ai.ChatMessage{Role: "user", Content: userContent})

    // Prepend system prompt
    messages := make([]ai.ChatMessage, 0, len(history)+1)
    messages = append(messages, ai.ChatMessage{Role: "system", Content: chatSystemPrompt})
    messages = append(messages, history...)

    var finalReply string
    var toolsUsed []string

    // Tool calling loop (tối đa 5 vòng)
    for i := 0; i < 5; i++ {
        resp, err := u.aiClient.Chat(ctx, messages, tools.AllTools)
        if err != nil {
            // Fallback về pattern nếu AI fail
            finalReply = u.patternResponse(req)
            break
        }
        if len(resp.Choices) == 0 {
            break
        }

        assistantMsg := resp.Choices[0].Message

        // AI không gọi tool → trả lời xong
        if len(assistantMsg.ToolCalls) == 0 {
            finalReply = assistantMsg.Content
            history = append(history, assistantMsg)
            break
        }

        // AI gọi tools → execute từng tool
        messages = append(messages, assistantMsg)

        for _, tc := range assistantMsg.ToolCalls {
            toolsUsed = append(toolsUsed, tc.Function.Name)

            result, err := u.executor.Execute(ctx, tc.Function.Name, tc.Function.Arguments)
            if err != nil {
                result = fmt.Sprintf("Tool error: %v", err)
            }

            messages = append(messages, ai.ChatMessage{
                Role:       "tool",
                Content:    result,
                ToolCallID: tc.ID,
                Name:       tc.Function.Name,
            })
        }
        // Tiếp tục loop — AI phân tích tool results và quyết định tiếp theo
    }

    if finalReply == "" {
        finalReply = "Xin lỗi, mình gặp sự cố khi xử lý câu hỏi này. Bạn thử hỏi lại nhé!"
    }

    // Lưu history (chỉ lưu user + assistant, không lưu tool messages)
    history = append(history, ai.ChatMessage{Role: "assistant", Content: finalReply})
    u.saveHistory(conversationID, history)

    return &dto.ChatResponse{
        Reply:          finalReply,
        ConversationID: conversationID,
        Provider:       "openai",
        ToolsUsed:      toolsUsed,
        ResponseTimeMs: time.Since(start).Milliseconds(),
    }, nil
}

func (u *chatbotUseCase) buildUserMessage(req *dto.ChatRequest) string {
    var sb strings.Builder
    sb.WriteString(req.Message)

    if req.ProblemID != nil {
        sb.WriteString(fmt.Sprintf("\n\n[Context: problem_id=%d", *req.ProblemID))
        if req.ProblemTitle != "" {
            sb.WriteString(fmt.Sprintf(", title=\"%s\"", req.ProblemTitle))
        }
        if req.UserID != nil {
            sb.WriteString(fmt.Sprintf(", user_id=%d", *req.UserID))
        }
        sb.WriteString("]")
    }

    if req.StudentSQL != "" {
        sb.WriteString(fmt.Sprintf("\n\n[Câu SQL của sinh viên:\n```sql\n%s\n```]", req.StudentSQL))
    }

    if req.ErrorMessage != "" {
        sb.WriteString(fmt.Sprintf("\n\n[Thông báo lỗi: %s]", req.ErrorMessage))
    }

    return sb.String()
}

func (u *chatbotUseCase) loadHistory(conversationID string) []ai.ChatMessage {
    if u.redis == nil {
        return []ai.ChatMessage{}
    }
    key := "chat_history:" + conversationID
    var history []ai.ChatMessage
    if err := u.redis.Get(key, &history); err != nil {
        return []ai.ChatMessage{}
    }
    // Giới hạn 20 messages gần nhất
    if len(history) > 20 {
        history = history[len(history)-20:]
    }
    return history
}

func (u *chatbotUseCase) saveHistory(conversationID string, history []ai.ChatMessage) {
    if u.redis == nil {
        return
    }
    key := "chat_history:" + conversationID
    _ = u.redis.SetWithExpiration(key, history, 2*time.Hour)
}

// patternResponse là fallback khi không có AI key
func (u *chatbotUseCase) patternResponse(req *dto.ChatRequest) string {
    msg := strings.ToLower(req.Message)
    errMsg := strings.ToLower(req.ErrorMessage)

    if strings.Contains(errMsg, "syntax error") {
        return "Câu SQL có lỗi cú pháp. Kiểm tra:\n1. Đủ dấu ngoặc () chưa?\n2. Thứ tự từ khóa SELECT → FROM → WHERE → GROUP BY → HAVING → ORDER BY?\n3. Dấu phẩy giữa các cột?"
    }
    if strings.Contains(errMsg, "does not exist") {
        return "Tên bảng hoặc cột không tồn tại. Kiểm tra chính tả và schema của bài."
    }
    if strings.Contains(msg, "join") {
        return "INNER JOIN chỉ lấy dòng khớp cả 2 bảng. LEFT JOIN lấy tất cả từ bảng trái.\nCú pháp: SELECT ... FROM a JOIN b ON a.id = b.a_id"
    }
    if strings.Contains(msg, "group by") {
        return "GROUP BY: mọi cột trong SELECT không trong hàm tổng hợp PHẢI có trong GROUP BY.\nVí dụ: SELECT dept, COUNT(*) FROM employees GROUP BY dept"
    }
    return "Hãy chia sẻ câu SQL của bạn để mình hỗ trợ cụ thể hơn!"
}
