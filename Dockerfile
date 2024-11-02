# Stage 1: Build
FROM golang:1.23.2-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if go.mod and go.sum are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
RUN go build -o ec2-restart-manager .

# Stage 2: Run
FROM alpine:latest

# Copy the binary from the builder stage
COPY --from=builder /app/ec2-restart-manager /usr/local/bin/ec2-restart-manager

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["ec2-restart-manager"]
