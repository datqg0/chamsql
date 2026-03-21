$ErrorActionPreference = "Stop"

Write-Host "Logging in as admin2..."
$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body '{"identifier": "admin2", "password": "123456"}' -ContentType "application/json"
$token = $response.data.accessToken

Write-Host "Fetching problems..."
$probs = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/problems" -Method Get
Write-Host "Count before update: $($probs.data.total)"

if ($probs.data.total -gt 0) {
    $id = $probs.data.problems[0].id
    Write-Host "Updating problem ID: $id"
    
    $body = '{"title":"Updated Prob","description":"This is a valid description with > 10 chars","difficulty":"easy","topicId":null,"supportedDatabases":["postgresql"],"orderMatters":false,"isPublic":true}'
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/problems/$id" -Method Put -Headers @{Authorization="Bearer $token"} -Body $body -ContentType "application/json" | Out-Null

    Write-Host "Fetching problems again..."
    $probsAfter = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/problems" -Method Get
    Write-Host "Count after update: $($probsAfter.data.total)"
} else {
    Write-Host "No problems to update."
}
