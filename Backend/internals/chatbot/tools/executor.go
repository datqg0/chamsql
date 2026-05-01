package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "strings"

    "backend/pkgs/runner"
    "backend/sql/models"
)

type ToolExecutor struct {
    queries *models.Queries
    runner  runner.Runner
}

func NewToolExecutor(queries *models.Queries, r runner.Runner) *ToolExecutor {
    return &ToolExecutor{queries: queries, runner: r}
}

func (e *ToolExecutor) Execute(ctx context.Context, toolName, argsJSON string) (string, error) {
    var args map[string]interface{}
    if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
        return "", fmt.Errorf("invalid args: %w", err)
    }

    switch toolName {
    case "get_problem_schema":
        return e.getProblemSchema(ctx, args)
    case "run_student_sql":
        return e.runStudentSQL(ctx, args)
    case "compare_with_solution":
        return e.compareWithSolution(ctx, args)
    case "get_student_history":
        return e.getStudentHistory(ctx, args)
    case "explain_sql_concept":
        return e.explainConcept(args)
    default:
        return fmt.Sprintf("unknown tool: %s", toolName), nil
    }
}

func (e *ToolExecutor) getProblemSchema(ctx context.Context, args map[string]interface{}) (string, error) {
    problemIDRaw, ok := args["problem_id"]
    if !ok {
        return "Missing problem_id", nil
    }
    problemID := int64(problemIDRaw.(float64))

    // Dùng query riêng — không bao giờ có SolutionQuery trong kết quả
    problem, err := e.queries.GetProblemSchemaForChatbot(ctx, problemID)
    if err != nil {
        return fmt.Sprintf("Không tìm thấy bài toán ID %d", problemID), nil
    }

    testCases, _ := e.queries.ListProblemTestCases(ctx, problemID)

    result := map[string]interface{}{
        "id":              problem.ID,
        "title":           problem.Title,
        "description":     problem.Description,
        "init_script":     problem.InitScript,
        "supported_dbs":   problem.SupportedDatabases,
        "difficulty":      problem.Difficulty,
        "test_case_count": len(testCases),
    }
    b, _ := json.MarshalIndent(result, "", "  ")
    return string(b), nil
}

func (e *ToolExecutor) runStudentSQL(ctx context.Context, args map[string]interface{}) (string, error) {
    sql, _ := args["sql"].(string)
    initScript, _ := args["init_script"].(string)

    // Auto-fetch init_script nếu rỗng nhưng có problem_id
    if initScript == "" {
        if pidRaw, ok := args["problem_id"]; ok {
            pid := int64(pidRaw.(float64))
            if p, err := e.queries.GetProblemSchemaForChatbot(ctx, pid); err == nil {
                initScript = p.InitScript
            }
        }
    }

    if initScript == "" {
        return "Cần init_script để chạy SQL. Hãy gọi get_problem_schema trước hoặc truyền problem_id.", nil
    }

    // Security guard — chỉ cho SELECT
    sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
    for _, kw := range []string{"DROP", "DELETE", "UPDATE", "INSERT", "TRUNCATE", "ALTER", "CREATE", "GRANT"} {
        if strings.Contains(sqlUpper, kw) {
            return "Chỉ hỗ trợ câu lệnh SELECT trong chế độ hỗ trợ học tập.", nil
        }
    }
    if !strings.HasPrefix(sqlUpper, "SELECT") && !strings.HasPrefix(sqlUpper, "WITH") {
        return "Câu SQL phải bắt đầu bằng SELECT hoặc WITH.", nil
    }

    result, err := e.runner.ExecuteWithSetup(ctx, runner.DBTypePostgreSQL, initScript, sql)
    if err != nil {
        return fmt.Sprintf("Lỗi khi chạy SQL: %v", err), nil
    }
    if result.Error != "" {
        return fmt.Sprintf("SQL Error: %s", result.Error), nil
    }

    b, _ := json.MarshalIndent(map[string]interface{}{
        "row_count":    len(result.Rows),
        "columns":      result.Columns,
        "rows":         result.Rows,
        "exec_time_ms": result.ExecutionMs,
    }, "", "  ")
    return string(b), nil
}

func (e *ToolExecutor) compareWithSolution(ctx context.Context, args map[string]interface{}) (string, error) {
    studentSQL, _ := args["student_sql"].(string)
    initScript, _ := args["init_script"].(string)
    problemIDRaw, ok := args["problem_id"]
    if !ok {
        return "Thiếu problem_id để thực hiện so sánh với đáp án.", nil
    }
    problemID := int64(problemIDRaw.(float64))

    // Lấy solution từ DB — không nhận từ AI caller để bảo mật
    problem, err := e.queries.GetProblemByID(ctx, problemID)
    if err != nil {
        return "Không tìm thấy bài toán.", nil
    }
    solutionSQL := problem.SolutionQuery
    if solutionSQL == "" {
        return "Bài toán này chưa có đáp án mẫu để so sánh.", nil
    }

    // Auto-fetch init_script nếu rỗng
    if initScript == "" {
        initScript = problem.InitScript
    }

    studentRes, sErr := e.runner.ExecuteWithSetup(ctx, runner.DBTypePostgreSQL, initScript, studentSQL)
    solutionRes, solErr := e.runner.ExecuteWithSetup(ctx, runner.DBTypePostgreSQL, initScript, solutionSQL)

    feedback := map[string]interface{}{}

    if sErr != nil || (studentRes != nil && studentRes.Error != "") {
        errMsg := ""
        if sErr != nil { errMsg = sErr.Error() }
        if studentRes != nil && studentRes.Error != "" { errMsg = studentRes.Error }
        feedback["has_error"] = true
        feedback["error"] = errMsg
        feedback["hint"] = "Câu SQL của bạn có lỗi. Hãy sửa lỗi này trước khi so sánh kết quả."
        b, _ := json.MarshalIndent(feedback, "", "  ")
        return string(b), nil
    }
    if solErr != nil {
        return "Lỗi hệ thống khi chạy đáp án mẫu.", nil
    }

    nStudent := len(studentRes.Rows)
    nSolution := len(solutionRes.Rows)
    feedback["student_rows"] = nStudent
    feedback["expected_rows"] = nSolution
    feedback["row_count_match"] = nStudent == nSolution

    if nStudent == nSolution {
        feedback["assessment"] = "Số dòng kết quả khớp với đáp án. Hãy kiểm tra thêm xem giá trị từng cột có đúng không."
    } else if nStudent < nSolution {
        feedback["assessment"] = fmt.Sprintf(
            "Kết quả thiếu %d dòng so với đáp án. Gợi ý: kiểm tra điều kiện WHERE, JOIN bị thiếu, hoặc cần dùng UNION.",
            nSolution-nStudent)
    } else {
        feedback["assessment"] = fmt.Sprintf(
            "Kết quả thừa %d dòng so với đáp án. Gợi ý: điều kiện WHERE chưa đủ chặt, JOIN tạo ra bản ghi trùng, hoặc cần DISTINCT.",
            nStudent-nSolution)
    }
    // Tuyệt đối không trả về rows của solution để tránh lộ đáp án
    b, _ := json.MarshalIndent(feedback, "", "  ")
    return string(b), nil
}

func (e *ToolExecutor) getStudentHistory(ctx context.Context, args map[string]interface{}) (string, error) {
    problemIDRaw, ok1 := args["problem_id"]
    userIDRaw, ok2 := args["user_id"]
    if !ok1 || !ok2 {
        return "Missing problem_id or user_id", nil
    }
    problemID := int64(problemIDRaw.(float64))
    userID := int64(userIDRaw.(float64))

    progress, err := e.queries.GetUserProblemProgress(ctx, models.GetUserProblemProgressParams{
        UserID:    userID,
        ProblemID: problemID,
    })
    if err != nil {
        return "Chưa có lịch sử submit cho bài này.", nil
    }

    b, _ := json.MarshalIndent(map[string]interface{}{
        "user_id":      userID,
        "problem_id":   problemID,
        "attempts":     progress.Attempts,
        "is_solved":    progress.IsSolved,
        "best_time_ms": progress.BestTimeMs,
    }, "", "  ")
    return string(b), nil
}

func (e *ToolExecutor) explainConcept(args map[string]interface{}) (string, error) {
    concept := strings.ToLower(fmt.Sprintf("%v", args["concept"]))

    kb := map[string]string{
        "join": "JOIN kết hợp dữ liệu từ nhiều bảng:\n• INNER JOIN: chỉ lấy dòng khớp cả 2 bảng\n• LEFT JOIN: tất cả từ bảng trái, bảng phải có thể NULL\n• RIGHT JOIN: ngược lại LEFT JOIN\n• FULL OUTER JOIN: tất cả 2 phía\nCú pháp: SELECT ... FROM a JOIN b ON a.id = b.a_id",
        "group by": "GROUP BY gom các dòng có giá trị giống nhau thành nhóm.\nQuy tắc vàng: mọi cột trong SELECT không nằm trong hàm tổng hợp PHẢI có trong GROUP BY.\nVí dụ: SELECT dept, COUNT(*) FROM employees GROUP BY dept",
        "having": "HAVING lọc kết quả SAU khi GROUP BY (WHERE chạy TRƯỚC group).\nVí dụ: SELECT dept, COUNT(*) FROM employees GROUP BY dept HAVING COUNT(*) > 5",
        "subquery": "Subquery là truy vấn lồng trong truy vấn khác:\n• Trong WHERE: WHERE id IN (SELECT id FROM ...)\n• Trong FROM: SELECT * FROM (SELECT ...) AS sub\n• Correlated subquery: tham chiếu bảng ngoài",
        "window function": "Window function tính toán trên cửa sổ dữ liệu mà không gom nhóm:\nROW_NUMBER(), RANK(), DENSE_RANK(), LAG(), LEAD()\nSUM(col) OVER (PARTITION BY x ORDER BY y)\nKhác aggregate: không làm mất dòng.",
        "cte": "CTE (Common Table Expression) tạo bảng tạm có tên:\nWITH ten_cte AS (SELECT ...) SELECT * FROM ten_cte\nCó thể tạo nhiều CTE: WITH a AS (...), b AS (...) SELECT ...\nRecursive CTE: WITH RECURSIVE cho dữ liệu cây/đệ quy.",
        "index": "Index giúp tăng tốc truy vấn bằng cách tạo cấu trúc tra cứu:\nCREATE INDEX idx_name ON table(column)\nNên index: cột WHERE thường dùng, cột JOIN, cột ORDER BY.\nKhông index quá nhiều: chậm INSERT/UPDATE.",
        "distinct": "DISTINCT loại bỏ dòng trùng lặp trong kết quả:\nSELECT DISTINCT column FROM table\nHoặc: COUNT(DISTINCT column) để đếm giá trị khác nhau.",
        "union": "UNION kết hợp kết quả từ 2 SELECT:\n• UNION: loại bỏ dòng trùng\n• UNION ALL: giữ tất cả kể cả trùng\nYêu cầu: cùng số cột và kiểu dữ liệu tương thích.",
    }

    for key, explanation := range kb {
        if strings.Contains(concept, key) || strings.Contains(key, concept) {
            return explanation, nil
        }
    }
    return fmt.Sprintf("Concept '%s' chưa có trong knowledge base — AI sẽ giải thích dựa trên kiến thức SQL chung.", concept), nil
}
