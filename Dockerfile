# --- Build Stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Cache modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build statically-linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-w -s -extldflags '-static'" -o /syac

# --- Runtime Stage ---
FROM alpine:latest

# Only include runtime deps
RUN apk add --no-cache docker-cli git

# Copy the built binary
COPY --from=builder /syac /usr/local/bin/syac

