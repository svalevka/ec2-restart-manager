# Use the official Golang image for building and running the app
FROM golang:1.23.2-alpine

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy all files from the current directory into the container
COPY . .

# Download all dependencies. Dependencies will be cached if go.mod and go.sum are not changed
RUN go mod download

# Build the Go app
RUN go build -o ec2-restart-manager .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./ec2-restart-manager"]
