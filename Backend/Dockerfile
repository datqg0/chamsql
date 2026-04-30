FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o main ./cmd/app/main.go

# Final stage
FROM alpine:latest

WORKDIR /root/

# Install runtime dependencies
RUN apk add --no-cache ca-certificates

# Copy the binary and config from builder
COPY --from=builder /app/main .
COPY --from=builder /app/.env .
# Optional: COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./main"]
