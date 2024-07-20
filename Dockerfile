# Start from a base image containing Go runtime
FROM golang:1.22 AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Build the Go app
RUN CGO_ENABLED=0 GOOS=linux go build -o ./main


# Start a new stage from scratch
FROM scratch

# Set the Current Working Directory inside the container
WORKDIR /app

# # Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main .
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/websockets.html .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["/app/main"]