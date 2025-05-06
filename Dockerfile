# Use the official Golang image for building and running the app
FROM golang:1.23.2-alpine
ARG VERSION="unknown"

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy all files from the current directory into the container
COPY . .

# Download all dependencies
RUN go mod download

# Build the Go app with explicit output path

RUN go build -ldflags="-X ec2-restart-manager/config.Version=${VERSION}" -o /app/bin/ec2-restart-manager .

# Create the bin directory if it doesn't exist
RUN mkdir -p /app/bin

# Set execution permissions on the binary
RUN chmod +x /app/bin/ec2-restart-manager

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable with full path
CMD ["/app/bin/ec2-restart-manager"]