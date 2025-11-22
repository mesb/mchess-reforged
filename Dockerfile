# --- Stage 1: Builder ---
FROM golang:1.23-alpine AS builder

# Set setup
WORKDIR /app

# Copy dependencies first (caching layer)
COPY go.mod ./
COPY go.sum ./

# Copy source code
COPY . .

# Build the server binary
# -o mchess-server: output filename
# ./cmd/server: entry point
RUN go build -o mchess-server ./cmd/server

# --- Stage 2: Runner ---
FROM alpine:latest

WORKDIR /root/

# Copy only the binary from the builder stage
COPY --from=builder /app/mchess-server .

# Expose the port defined in main.go
EXPOSE 8080

# Command to run when container starts
CMD ["./mchess-server"]
