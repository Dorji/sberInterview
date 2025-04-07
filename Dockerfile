# Build stage
FROM golang:1.24.1-alpine AS builder

# Set working directory
WORKDIR /app

RUN mkdir config

# Install dependencies for yaml
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code and config
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/sber_loan ./cmd/sber_loan

# Final stage
FROM alpine:3.19

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/bin/sber_loan /app/sber_loan
# Copy config file
COPY --from=builder /app/config/config.yml /app/config/config.yml

# Expose both ports (можно переопределить через config.yml)
EXPOSE 8080 50051

# Command to run the application
CMD ["/app/sber_loan"]