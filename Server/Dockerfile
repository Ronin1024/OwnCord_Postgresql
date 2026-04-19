# Multi-stage build for image size optimization
FROM golang:1.26.1-alpine AS builder

# Installing required packages for building
RUN apk add --no-cache git ca-certificates tzdata

# Creating a user for security
RUN adduser -D -g '' owncord owncord

# Setting working directory
WORKDIR /build

# Copying go.mod and go.sum for dependency caching
COPY go.mod go.sum ./

# Downloading dependencies
RUN go mod download
RUN go mod verify

# Copying source code
COPY . .

# Compiling
RUN  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o owncord -ldflags "-s -w -X main.version=1.0.0" .

# Final image
FROM alpine:latest

# Installing required packages
RUN apk --no-cache add ca-certificates tzdata

# Copying user from builder
COPY --from=builder /etc/passwd /etc/passwd

# Creating application directory
WORKDIR /app

RUN export HOME=/tmp

# Copying compiled application
COPY --from=builder /build/owncord .

# Copying livekit
COPY --from=builder /build/livekit ./livekit

# Copying migrations
COPY --from=builder /build/migrations ./migrations

# Setup user permissions
RUN chown -R owncord /app
RUN chmod +x /app/livekit/livekit-server

# Switching to unprivileged user
USER owncord

# Opening ports
EXPOSE 80
EXPOSE 8443
EXPOSE 7880
EXPOSE 7881
EXPOSE 50000-60000/udp

# Application health check. Change localhost to your host
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=2 \
    CMD wget --no-verbose --tries=1 --spider https://localhost:8443/health || exit 1

# Starting the application
CMD ["./owncord"]
