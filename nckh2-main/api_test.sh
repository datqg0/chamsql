# SQL Exam API - cURL Test Commands
# Base URL: http://localhost:8080/api/v1

# =============================================
# AUTH ENDPOINTS
# =============================================

# 1. Register
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "student@test.com",
    "username": "student1",
    "password": "123456",
    "fullName": "Test Student",
    "studentId": "SV001"
  }'

# 2. Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "student@test.com",
    "password": "123456"
  }'

# Save the token from login response:
# export TOKEN="your_access_token_here"

# =============================================
# TOPICS ENDPOINTS
# =============================================

# 3. List Topics (public)
curl -X GET http://localhost:8080/api/v1/topics

# 4. Get Topic by Slug (public)
curl -X GET http://localhost:8080/api/v1/topics/basic-select

# 5. Create Topic (lecturer/admin only)
curl -X POST http://localhost:8080/api/v1/topics \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "name": "Basic SELECT",
    "slug": "basic-select",
    "description": "Learn basic SELECT queries",
    "icon": "📊",
    "sortOrder": 1
  }'

# =============================================
# PROBLEMS ENDPOINTS
# =============================================

# 6. List Problems (public)
curl -X GET "http://localhost:8080/api/v1/problems?page=1&pageSize=10"

# 7. List Problems by Difficulty
curl -X GET "http://localhost:8080/api/v1/problems?difficulty=easy"

# 8. Get Problem by Slug (public)
curl -X GET http://localhost:8080/api/v1/problems/simple-select

# 9. Create Problem (lecturer/admin only)
curl -X POST http://localhost:8080/api/v1/problems \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "Simple SELECT",
    "slug": "simple-select",
    "description": "Select all employees from the employees table",
    "difficulty": "easy",
    "initScript": "",
    "solutionQuery": "SELECT * FROM employees",
    "supportedDatabases": ["postgresql", "mysql"],
    "orderMatters": false,
    "isPublic": true
  }'

# =============================================
# SUBMISSIONS ENDPOINTS
# =============================================

# 10. Run Query (test without submitting - public)
curl -X POST http://localhost:8080/api/v1/problems/1/run \
  -H "Content-Type: application/json" \
  -d '{
    "code": "SELECT * FROM employees",
    "databaseType": "postgresql"
  }'

# 11. Submit Solution (auth required)
curl -X POST http://localhost:8080/api/v1/problems/1/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "code": "SELECT * FROM employees",
    "databaseType": "postgresql"
  }'

# 12. List My Submissions (auth required)
curl -X GET "http://localhost:8080/api/v1/submissions?page=1&pageSize=10" \
  -H "Authorization: Bearer $TOKEN"

# =============================================
# EXAMS ENDPOINTS
# =============================================

# 13. List Exams (auth required)
curl -X GET http://localhost:8080/api/v1/exams \
  -H "Authorization: Bearer $TOKEN"

# 14. Create Exam (lecturer/admin only)
curl -X POST http://localhost:8080/api/v1/exams \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "title": "SQL Basics Exam",
    "description": "Test your SQL basics",
    "startTime": "2026-01-19T08:00:00Z",
    "endTime": "2026-01-19T10:00:00Z",
    "durationMinutes": 60,
    "allowedDatabases": ["postgresql"],
    "allowAiAssistance": false,
    "shuffleProblems": true,
    "showResultImmediately": true,
    "maxAttempts": 3,
    "isPublic": true
  }'

# 15. Add Problem to Exam (lecturer/admin)
curl -X POST http://localhost:8080/api/v1/exams/1/problems \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "problemId": 1,
    "points": 10,
    "sortOrder": 1
  }'

# 16. Add Participants to Exam (lecturer/admin)
curl -X POST http://localhost:8080/api/v1/exams/1/participants \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "userIds": [1, 2, 3]
  }'

# 17. Start Exam (student)
curl -X POST http://localhost:8080/api/v1/exams/1/start \
  -H "Authorization: Bearer $TOKEN"

# 18. Submit Answer in Exam (student)
curl -X POST http://localhost:8080/api/v1/exams/1/submit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "problemId": 1,
    "code": "SELECT * FROM employees",
    "databaseType": "postgresql"
  }'

# 19. Finish Exam (student)
curl -X POST http://localhost:8080/api/v1/exams/1/finish \
  -H "Authorization: Bearer $TOKEN"

# 20. Get My Exams (student)
curl -X GET http://localhost:8080/api/v1/my-exams \
  -H "Authorization: Bearer $TOKEN"

# =============================================
# ADMIN ENDPOINTS (admin only)
# =============================================

# 21. Get System Stats
curl -X GET http://localhost:8080/api/v1/admin/stats \
  -H "Authorization: Bearer $TOKEN"

# 22. List Users
curl -X GET "http://localhost:8080/api/v1/admin/users?page=1&pageSize=20" \
  -H "Authorization: Bearer $TOKEN"

# 23. Import Users
curl -X POST http://localhost:8080/api/v1/admin/users/import \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "users": [
      {
        "email": "student2@test.com",
        "username": "student2",
        "fullName": "Student Two",
        "studentId": "SV002",
        "role": "student"
      },
      {
        "email": "lecturer@test.com",
        "username": "lecturer1",
        "fullName": "Lecturer One",
        "role": "lecturer"
      }
    ]
  }'

# 24. Update User Role
curl -X PUT http://localhost:8080/api/v1/admin/users/2/role \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "role": "lecturer"
  }'

# =============================================
# HEALTH CHECK
# =============================================

# 25. Health Check
curl -X GET http://localhost:8080/health
