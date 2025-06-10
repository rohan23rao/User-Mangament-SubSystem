FROM golang:1.21-alpine AS development

# Install necessary packages for development
RUN apk add --no-cache git curl postgresql-client

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install air for live reloading in development
RUN go install github.com/cosmtrek/air@latest

# Copy source code
COPY . .

# Expose port
EXPOSE 3000

# Default command for development (can be overridden)
CMD ["air", "-c", ".air.toml"]

# Production stage
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM alpine:latest AS production
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 3000
CMD ["./main"]