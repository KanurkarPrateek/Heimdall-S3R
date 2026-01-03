# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN go build -o rpc-load-balancer ./cmd/server

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/rpc-load-balancer .
COPY --from=builder /app/config ./config
COPY --from=builder /app/.env .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./rpc-load-balancer"]
