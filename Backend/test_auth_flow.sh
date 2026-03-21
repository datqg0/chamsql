#!/bin/bash

BASE_URL="http://localhost:8080"

# 1. Login
echo "1. Logging in..."
LOGIN_RESPONSE=$(curl -s -c cookies.txt -i -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier": "student1", "password": "password123"}')

# Extract Access Token (simple grep/sed, assuming JSON response)
ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
    echo "Login failed. Response:"
    echo "$LOGIN_RESPONSE"
    exit 1
fi

echo "Login success. Access Token: ${ACCESS_TOKEN:0:20}..."

# 2. Access Protected Route (e.g., Get Me or Dashboard stats)
# We don't have a direct /me endpoint in the snippet, let's try /admin/users (need admin) or just check header/cookie
# Let's try to logout (requires auth) but verify with a fake token first to see 401
echo "2. Accessing Protected Route (Logout) with VALID token..."
LOGOUT_RESPONSE=$(curl -s -b cookies.txt -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

echo "Logout Response: $LOGOUT_RESPONSE"
# Expect success message

# 3. Validation: Refresh Token
# Login again to get fresh tokens
echo "3. Logging in again for Refresh Test..."
curl -s -c cookies.txt -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"identifier": "student1", "password": "password123"}' > /dev/null

echo "4. Refreshing Token..."
REFRESH_RESPONSE=$(curl -s -b cookies.txt -c cookies_new.txt -X POST "$BASE_URL/auth/refresh")

echo "Refresh Response: $REFRESH_RESPONSE"
# Check if we got new access token
NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | grep -o '"accessToken":"[^"]*"' | cut -d'"' -f4)

if [ -z "$NEW_ACCESS_TOKEN" ]; then
    echo "Refresh failed."
    exit 1
fi

echo "Refresh success. New Access Token: ${NEW_ACCESS_TOKEN:0:20}..."
