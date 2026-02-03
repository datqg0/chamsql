# 1. REGISTER
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "0987654321",
    "username": "nguyenvana",
    "password": "password123",
    "fullName": "Nguyen Van A",
    "email": "nguyenvana@example.com"
  }'

# 2. LOGIN (Using Phone) - Save Cookies
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "0987654321",
    "password": "password123"
  }'

# 2. LOGIN (Using Username) - Save Cookies
curl -c cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "nguyenvana",
    "password": "password123"
  }'

# 3. REFRESH TOKEN (Use Cookie)
# No body needed if cookie is present
curl -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{}'

# 4. LOGOUT
# Replace <ACCESS_TOKEN>
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer <ACCESS_TOKEN>"
