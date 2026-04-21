# Test Script for ChamSQL Exam Flow
# Run this script to test the complete end-to-end flow

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$LecturerToken = "",
    [string]$StudentToken = ""
)

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "ChamSQL Exam Flow Test Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Helper function for API calls
function Invoke-API {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [object]$Body = $null
    )
    
    $url = "$BaseUrl/api/v1$Endpoint"
    $params = @{
        Uri = $url
        Method = $Method
        Headers = $Headers
    }
    
    if ($Body) {
        $params.Body = ($Body | ConvertTo-Json -Depth 10)
        $params.ContentType = "application/json"
    }
    
    try {
        $response = Invoke-RestMethod @params
        return @{ Success = $true; Data = $response }
    } catch {
        return @{ Success = $false; Error = $_.Exception.Message }
    }
}

# Test 1: Health Check
Write-Host "[Test 1] Health Check..." -ForegroundColor Yellow
$health = Invoke-API -Method "GET" -Endpoint "/health"
if ($health.Success) {
    Write-Host "  Health: OK" -ForegroundColor Green
} else {
    Write-Host "  Health: FAIL - $($health.Error)" -ForegroundColor Red
}
Write-Host ""

# Test 2: Sandbox Status (Admin/Lecturer)
Write-Host "[Test 2] Sandbox Status..." -ForegroundColor Yellow
if ($LecturerToken) {
    $headers = @{ "Authorization" = "Bearer $LecturerToken" }
    $sandbox = Invoke-API -Method "GET" -Endpoint "/admin/sandbox/status" -Headers $headers
    if ($sandbox.Success) {
        Write-Host "  PostgreSQL: $($sandbox.Data.postgres.connected)" -ForegroundColor $(if ($sandbox.Data.postgres.connected) { "Green" } else { "Red" })
        Write-Host "  MySQL: $($sandbox.Data.mysql.connected)" -ForegroundColor $(if ($sandbox.Data.mysql.connected) { "Green" } else { "Red" })
        Write-Host "  SQLServer: $($sandbox.Data.sqlserver.connected)" -ForegroundColor $(if ($sandbox.Data.sqlserver.connected) { "Green" } else { "Red" })
    } else {
        Write-Host "  Sandbox: FAIL - $($sandbox.Error)" -ForegroundColor Red
    }
} else {
    Write-Host "  Skipped (no lecturer token)" -ForegroundColor Gray
}
Write-Host ""

# Test 3: Create Exam (Lecturer)
Write-Host "[Test 3] Create Exam..." -ForegroundColor Yellow
if ($LecturerToken) {
    $headers = @{ "Authorization" = "Bearer $LecturerToken" }
    $startTime = (Get-Date).AddHours(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
    $endTime = (Get-Date).AddHours(3).ToString("yyyy-MM-ddTHH:mm:ssZ")
    
    $examBody = @{
        title = "Test Exam - $(Get-Date -Format 'yyyyMMddHHmm')"
        description = "Test exam created by automated script"
        start_time = $startTime
        end_time = $endTime
        duration_minutes = 120
        max_attempts = 1
        is_public = $false
        allowed_databases = @("postgresql", "mysql")
    }
    
    $exam = Invoke-API -Method "POST" -Endpoint "/exams" -Headers $headers -Body $examBody
    if ($exam.Success) {
        $examId = $exam.Data.id
        Write-Host "  Created Exam ID: $examId" -ForegroundColor Green
        
        # Test 4: Add Problem to Exam
        Write-Host "[Test 4] Add Problem to Exam..." -ForegroundColor Yellow
        $problemBody = @{
            problem_id = 1  # Assuming problem ID 1 exists
            points = 10
            sort_order = 1
        }
        $problem = Invoke-API -Method "POST" -Endpoint "/exams/$examId/problems" -Headers $headers -Body $problemBody
        if ($problem.Success) {
            Write-Host "  Problem added successfully" -ForegroundColor Green
        } else {
            Write-Host "  Add Problem: FAIL - $($problem.Error)" -ForegroundColor Red
        }
        
        # Test 5: Add Participant
        Write-Host "[Test 5] Add Participant..." -ForegroundColor Yellow
        # This would need actual student user IDs
        Write-Host "  Skipped (requires student user IDs)" -ForegroundColor Gray
    } else {
        Write-Host "  Create Exam: FAIL - $($exam.Error)" -ForegroundColor Red
    }
} else {
    Write-Host "  Skipped (no lecturer token)" -ForegroundColor Gray
}
Write-Host ""

# Test 6: List Exams (Student)
Write-Host "[Test 6] List My Exams (Student)..." -ForegroundColor Yellow
if ($StudentToken) {
    $headers = @{ "Authorization" = "Bearer $StudentToken" }
    $myExams = Invoke-API -Method "GET" -Endpoint "/my-exams" -Headers $headers
    if ($myExams.Success) {
        Write-Host "  Found $($myExams.Data.exams.Count) exams" -ForegroundColor Green
    } else {
        Write-Host "  List Exams: FAIL - $($myExams.Error)" -ForegroundColor Red
    }
} else {
    Write-Host "  Skipped (no student token)" -ForegroundColor Gray
}
Write-Host ""

# Test 7: PDF Upload (Lecturer)
Write-Host "[Test 7] PDF Upload..." -ForegroundColor Yellow
Write-Host "  Note: Test with actual PDF file via frontend" -ForegroundColor Gray
Write-Host "  API Endpoint: POST /api/v1/lecturer/pdf/upload" -ForegroundColor Gray
Write-Host ""

# Summary
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl" -ForegroundColor White
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Start backend: cd Backend && go run cmd/app/main.go" -ForegroundColor White
Write-Host "2. Start frontend: cd Frontend && npm run dev" -ForegroundColor White
Write-Host "3. Login as lecturer and get token" -ForegroundColor White
Write-Host "4. Run this script with tokens:" -ForegroundColor White
Write-Host "   .\test_exam_flow.ps1 -LecturerToken 'your_token' -StudentToken 'student_token'" -ForegroundColor White
Write-Host ""
Write-Host "Manual Testing Checklist:" -ForegroundColor Yellow
Write-Host "[ ] Upload PDF with problems" -ForegroundColor White
Write-Host "[ ] Review extracted problems" -ForegroundColor White
Write-Host "[ ] Input solution queries for each problem" -ForegroundColor White
Write-Host "[ ] Create exam and add problems" -ForegroundColor White
Write-Host "[ ] Student start exam" -ForegroundColor White
Write-Host "[ ] Student submit SQL answers" -ForegroundColor White
Write-Host "[ ] Auto-grading execution" -ForegroundColor White
Write-Host "[ ] Lecturer manual grading" -ForegroundColor White
Write-Host "[ ] View grading results" -ForegroundColor White
Write-Host ""
