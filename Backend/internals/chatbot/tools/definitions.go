package tools

import "backend/pkgs/ai"

var GetProblemSchemaTool = ai.ChatTool{
    Type: "function",
    Function: ai.ChatToolFunction{
        Name:        "get_problem_schema",
        Description: "Lấy thông tin schema (cấu trúc bảng, dữ liệu mẫu, init_script) của bài toán SQL mà sinh viên đang làm. Gọi tool này trước khi tư vấn về bài cụ thể.",
        Parameters: ai.ToolParameters{
            Type: "object",
            Properties: map[string]ai.Property{
                "problem_id": {Type: "integer", Description: "ID của bài toán"},
            },
            Required: []string{"problem_id"},
        },
    },
}

var RunStudentSQLTool = ai.ChatTool{
    Type: "function",
    Function: ai.ChatToolFunction{
        Name:        "run_student_sql",
        Description: "Chạy thử câu SQL của sinh viên trong sandbox an toàn để xem kết quả thực tế hoặc thông báo lỗi. Chỉ hỗ trợ câu lệnh SELECT.",
        Parameters: ai.ToolParameters{
            Type: "object",
            Properties: map[string]ai.Property{
                "sql":         {Type: "string", Description: "Câu SQL SELECT cần chạy thử"},
                "init_script": {Type: "string", Description: "Script khởi tạo dữ liệu của bài (từ get_problem_schema)"},
            },
            Required: []string{"sql", "init_script"},
        },
    },
}

var CompareWithSolutionTool = ai.ChatTool{
	Type: "function",
	Function: ai.ChatToolFunction{
		Name:        "compare_with_solution",
		Description: "So sánh kết quả câu SQL của sinh viên với đáp án mẫu. Cần problem_id để lấy đáp án từ hệ thống. Trả về nhận xét về sự khác biệt mà KHÔNG tiết lộ nội dung đáp án.",
		Parameters: ai.ToolParameters{
			Type: "object",
			Properties: map[string]ai.Property{
				"problem_id":  {Type: "integer", Description: "ID bài toán để lấy đáp án từ hệ thống"},
				"student_sql": {Type: "string", Description: "Câu SQL của sinh viên"},
				"init_script": {Type: "string", Description: "Script khởi tạo dữ liệu của bài (có thể để trống nếu đã gọi get_problem_schema)"},
			},
			Required: []string{"problem_id", "student_sql"},
		},
	},
}

var GetStudentHistoryTool = ai.ChatTool{
    Type: "function",
    Function: ai.ChatToolFunction{
        Name:        "get_student_history",
        Description: "Xem lịch sử các lần submit của sinh viên cho bài này: đã thử bao nhiêu lần, lần nào đúng, lần nào sai.",
        Parameters: ai.ToolParameters{
            Type: "object",
            Properties: map[string]ai.Property{
                "problem_id": {Type: "integer", Description: "ID bài toán"},
                "user_id":    {Type: "integer", Description: "ID sinh viên"},
            },
            Required: []string{"problem_id", "user_id"},
        },
    },
}

var ExplainConceptTool = ai.ChatTool{
    Type: "function",
    Function: ai.ChatToolFunction{
        Name:        "explain_sql_concept",
        Description: "Lấy giải thích chi tiết về một SQL concept cụ thể từ knowledge base (JOIN, GROUP BY, subquery, window function, CTE, HAVING, ...).",
        Parameters: ai.ToolParameters{
            Type: "object",
            Properties: map[string]ai.Property{
                "concept": {Type: "string", Description: "Tên concept SQL cần giải thích"},
            },
            Required: []string{"concept"},
        },
    },
}

var AllTools = []ai.ChatTool{
    GetProblemSchemaTool,
    RunStudentSQLTool,
    CompareWithSolutionTool,
    GetStudentHistoryTool,
    ExplainConceptTool,
}
