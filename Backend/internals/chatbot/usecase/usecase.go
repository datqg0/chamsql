package usecase

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"backend/configs"
	"backend/internals/chatbot/controller/dto"

	"github.com/google/uuid"
)

// IChatbotUseCase defines chatbot operations for student guidance
type IChatbotUseCase interface {
	Ask(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error)
}

type chatbotUseCase struct {
	cfg        *configs.Config
	httpClient *http.Client
}

// NewChatbotUseCase creates a new chatbot usecase
func NewChatbotUseCase(cfg *configs.Config) IChatbotUseCase {
	return &chatbotUseCase{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Ask processes a student question and returns guidance
func (u *chatbotUseCase) Ask(ctx context.Context, req *dto.ChatRequest) (*dto.ChatResponse, error) {
	startTime := time.Now()

	// Generate conversation ID if not provided (multi-turn)
	conversationID := req.ConversationID
	if conversationID == "" {
		conversationID = uuid.New().String()
	}

	// Build the prompt with problem context
	prompt := u.buildPrompt(req)

	// Try HuggingFace API first
	if u.cfg.HuggingFaceAPIKey != "" {
		reply, err := u.callHuggingFace(ctx, prompt)
		if err == nil && reply != "" {
			suggestions := u.generateSuggestions(req)
			return &dto.ChatResponse{
				Reply:          reply,
				Suggestions:    suggestions,
				Hints:          u.generateHints(req),
				ConversationID: conversationID,
				Provider:       "huggingface",
				ResponseTimeMs: time.Since(startTime).Milliseconds(),
			}, nil
		}
	}

	// Fallback to pattern-based response
	reply := u.patternResponse(req)
	return &dto.ChatResponse{
		Reply:          reply,
		Suggestions:    u.generateSuggestions(req),
		Hints:          u.generateHints(req),
		ConversationID: conversationID,
		Provider:       "pattern",
		ResponseTimeMs: time.Since(startTime).Milliseconds(),
	}, nil
}

// buildPrompt constructs the AI prompt with problem context
func (u *chatbotUseCase) buildPrompt(req *dto.ChatRequest) string {
	var sb strings.Builder

	sb.WriteString("Bạn là một trợ lý hướng dẫn SQL thông minh cho sinh viên. ")
	sb.WriteString("Hãy giúp sinh viên hiểu cách viết câu truy vấn SQL đúng, ")
	sb.WriteString("nhưng KHÔNG đưa ra lời giải trực tiếp. Hướng dẫn từng bước.\n\n")

	if req.ProblemTitle != "" {
		sb.WriteString(fmt.Sprintf("Bài toán: %s\n", req.ProblemTitle))
	}
	if req.ProblemDesc != "" {
		sb.WriteString(fmt.Sprintf("Mô tả: %s\n", req.ProblemDesc))
	}
	if req.StudentSQL != "" {
		sb.WriteString(fmt.Sprintf("Câu SQL của sinh viên:\n```sql\n%s\n```\n", req.StudentSQL))
	}
	if req.ErrorMessage != "" {
		sb.WriteString(fmt.Sprintf("Lỗi gặp phải: %s\n", req.ErrorMessage))
	}

	sb.WriteString(fmt.Sprintf("\nCâu hỏi của sinh viên: %s\n", req.Message))
	sb.WriteString("\nHãy trả lời bằng tiếng Việt, hướng dẫn cụ thể nhưng không cho đáp án.")

	return sb.String()
}

// callHuggingFace calls HuggingFace Inference API
func (u *chatbotUseCase) callHuggingFace(ctx context.Context, prompt string) (string, error) {
	// Use a conversational model — Mistral or similar
	apiURL := "https://api-inference.huggingface.co/models/mistralai/Mistral-7B-Instruct-v0.3"

	payload := map[string]interface{}{
		"inputs": prompt,
		"parameters": map[string]interface{}{
			"max_new_tokens":  512,
			"temperature":     0.7,
			"top_p":           0.9,
			"return_full_text": false,
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(body))
	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Authorization", "Bearer "+u.cfg.HuggingFaceAPIKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := u.httpClient.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("HuggingFace API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result []struct {
		GeneratedText string `json:"generated_text"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result) == 0 || result[0].GeneratedText == "" {
		return "", fmt.Errorf("empty response from HuggingFace")
	}

	return result[0].GeneratedText, nil
}

// patternResponse generates a helpful response based on error patterns
func (u *chatbotUseCase) patternResponse(req *dto.ChatRequest) string {
	msg := strings.ToLower(req.Message)
	errMsg := strings.ToLower(req.ErrorMessage)
	sql := strings.ToLower(req.StudentSQL)

	// Error-based patterns
	if strings.Contains(errMsg, "syntax error") {
		return "Câu SQL của bạn có lỗi cú pháp. Hãy kiểm tra:\n" +
			"1. Đã đóng đủ dấu ngoặc () chưa?\n" +
			"2. Các từ khóa SQL (SELECT, FROM, WHERE...) có đúng thứ tự không?\n" +
			"3. Tên bảng và cột có chính xác không?\n" +
			"4. Dấu phẩy giữa các cột đã đủ chưa?"
	}

	if strings.Contains(errMsg, "column") && strings.Contains(errMsg, "does not exist") {
		return "Cột bạn sử dụng không tồn tại trong bảng. Hãy kiểm tra:\n" +
			"1. Xem lại tên cột có đúng chính tả không\n" +
			"2. Nếu JOIN nhiều bảng, hãy dùng alias (VD: t.column_name)\n" +
			"3. Kiểm tra xem cột đó thuộc bảng nào"
	}

	if strings.Contains(errMsg, "relation") && strings.Contains(errMsg, "does not exist") {
		return "Tên bảng bạn sử dụng không tồn tại. Hãy kiểm tra:\n" +
			"1. Tên bảng có đúng chính tả không?\n" +
			"2. Có đang dùng đúng schema không?\n" +
			"3. Bảng có thể có tên khác (số nhiều/ít)"
	}

	if strings.Contains(errMsg, "timeout") {
		return "Câu truy vấn bị timeout. Gợi ý:\n" +
			"1. Kiểm tra xem có vòng lặp vô hạn (subquery đệ quy) không\n" +
			"2. Thêm LIMIT để giới hạn kết quả\n" +
			"3. Sử dụng WHERE để lọc dữ liệu trước khi xử lý\n" +
			"4. Tránh SELECT * khi không cần thiết"
	}

	if strings.Contains(errMsg, "group by") || strings.Contains(errMsg, "aggregate") {
		return "Lỗi liên quan đến GROUP BY. Hãy nhớ:\n" +
			"1. Mọi cột trong SELECT không nằm trong hàm tổng hợp (SUM, COUNT...) phải có trong GROUP BY\n" +
			"2. Nếu dùng HAVING, nó phải đi sau GROUP BY\n" +
			"3. Không thể dùng alias trong WHERE, dùng trong HAVING thay thế"
	}

	// Query pattern suggestions
	if strings.Contains(msg, "join") || strings.Contains(msg, "nối") {
		return "Về JOIN trong SQL:\n" +
			"- INNER JOIN: Chỉ lấy các bản ghi có trong cả 2 bảng\n" +
			"- LEFT JOIN: Lấy tất cả từ bảng trái, bảng phải có thể NULL\n" +
			"- RIGHT JOIN: Ngược lại LEFT JOIN\n" +
			"Ví dụ: SELECT a.name, b.score FROM students a JOIN scores b ON a.id = b.student_id\n\n" +
			"Hãy kiểm tra điều kiện ON có đúng khóa ngoại không."
	}

	if strings.Contains(msg, "subquery") || strings.Contains(msg, "truy vấn con") {
		return "Subquery (truy vấn con) có 3 cách dùng chính:\n" +
			"1. Trong WHERE: WHERE column IN (SELECT ...)\n" +
			"2. Trong FROM: SELECT * FROM (SELECT ...) AS sub\n" +
			"3. Trong SELECT: SELECT (SELECT COUNT(*) FROM ...) AS cnt\n\n" +
			"Lưu ý: Subquery trong WHERE nên trả về ít dòng để tối ưu hiệu suất."
	}

	if strings.Contains(msg, "aggregate") || strings.Contains(msg, "tổng hợp") || strings.Contains(msg, "count") || strings.Contains(msg, "sum") {
		return "Các hàm tổng hợp (Aggregate Functions):\n" +
			"- COUNT(*): Đếm số dòng\n" +
			"- SUM(column): Tính tổng\n" +
			"- AVG(column): Tính trung bình\n" +
			"- MAX(column) / MIN(column): Lớn nhất / Nhỏ nhất\n\n" +
			"Nhớ dùng GROUP BY khi kết hợp cột thường và hàm tổng hợp."
	}

	// SQL-related hints
	if sql != "" && !strings.Contains(sql, "select") {
		return "Câu truy vấn SQL cần bắt đầu bằng SELECT. Cấu trúc cơ bản:\n" +
			"SELECT cột1, cột2 FROM tên_bảng WHERE điều_kiện\n\n" +
			"Bạn hãy thử viết lại câu truy vấn theo cấu trúc này."
	}

	// Default helpful response
	return "Mình hiểu câu hỏi của bạn. Một số gợi ý chung:\n" +
		"1. Đọc kỹ mô tả bài toán để hiểu yêu cầu\n" +
		"2. Xác định bảng nào cần dùng và mối quan hệ giữa chúng\n" +
		"3. Viết SELECT cơ bản trước, sau đó thêm điều kiện\n" +
		"4. Dùng nút 'Run' để test trước khi Submit\n" +
		"5. Nếu bị lỗi, đọc kỹ thông báo lỗi — nó thường chỉ rõ vấn đề\n\n" +
		"Bạn có thể gửi câu SQL cụ thể để mình hướng dẫn chi tiết hơn!"
}

// generateSuggestions generates quick-reply suggestions based on context
func (u *chatbotUseCase) generateSuggestions(req *dto.ChatRequest) []string {
	suggestions := []string{}

	if req.ErrorMessage != "" {
		suggestions = append(suggestions, "Giải thích lỗi này")
		suggestions = append(suggestions, "Sửa lỗi cú pháp")
	}

	if req.StudentSQL != "" {
		suggestions = append(suggestions, "Review câu SQL của tôi")
		suggestions = append(suggestions, "Tối ưu truy vấn")
	}

	if req.ProblemID != nil {
		suggestions = append(suggestions, "Gợi ý cách tiếp cận")
		suggestions = append(suggestions, "Giải thích yêu cầu bài")
	}

	// Always include general suggestions
	suggestions = append(suggestions, "Hướng dẫn JOIN")
	suggestions = append(suggestions, "Hướng dẫn GROUP BY")

	return suggestions
}

// generateHints generates progressive hints (not giving away the answer)
func (u *chatbotUseCase) generateHints(req *dto.ChatRequest) []string {
	hints := []string{}

	if req.ProblemDesc != "" {
		desc := strings.ToLower(req.ProblemDesc)

		if strings.Contains(desc, "join") || strings.Contains(desc, "kết hợp") {
			hints = append(hints, "Bài này cần JOIN các bảng. Xác định khóa ngoại để nối.")
		}
		if strings.Contains(desc, "count") || strings.Contains(desc, "đếm") {
			hints = append(hints, "Sử dụng COUNT() kết hợp GROUP BY.")
		}
		if strings.Contains(desc, "max") || strings.Contains(desc, "lớn nhất") {
			hints = append(hints, "Dùng MAX() hoặc ORDER BY ... DESC LIMIT 1.")
		}
		if strings.Contains(desc, "average") || strings.Contains(desc, "trung bình") {
			hints = append(hints, "Sử dụng AVG() để tính trung bình.")
		}
	}

	return hints
}
