FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o modelplex ./cmd/modelplex

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/modelplex .

# Copy default config
COPY config.toml .

# Create directory for socket
RUN mkdir -p /tmp

# Expose the socket directory as volume
VOLUME ["/tmp"]

ENTRYPOINT ["./modelplex"]
CMD ["--config", "config.toml", "--socket", "/tmp/modelplex.socket"]