# =========================
# BUILD STAGE
# =========================
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git (untuk go mod)
RUN apk add --no-cache git

# Copy go files
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o rt-management main.go

# =========================
# RUNTIME STAGE
# =========================
FROM alpine:3.19

WORKDIR /app

# SSL cert (penting untuk HTTPS, DB, dll)
RUN apk add --no-cache ca-certificates

# Copy binary
COPY --from=builder /app/rt-management .

# Copy env (opsional, biasanya pakai --env-file)
# COPY .env .

EXPOSE 8080

CMD ["./rt-management"]
