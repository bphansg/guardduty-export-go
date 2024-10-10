# Build stage
FROM golang:1.23 AS builder

WORKDIR /app

# Set GOPROXY to direct
ENV GOPROXY=direct

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the application with all dependencies statically linked
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Final stage
FROM debian:buster-slim

WORKDIR /app

# Copy the pre-built binary file from the previous stage
COPY --from=builder /app/main .

# Copy CA certificates from the builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]

