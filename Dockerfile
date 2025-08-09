# Stage 1: Build the Go application
FROM golang:tip-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
# CGO_ENABLED=0 is used to build a statically linked binary
# -o /app/chat-server specifies the output file name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app/chat-server .

# Stage 2: Create a lightweight image
FROM alpine:latest

# Add ca-certificates to trust TLS certificates
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/chat-server .

# Expose port 4545 to the outside world
EXPOSE 4545

# Command to run the executable
CMD ["./chat-server"]
