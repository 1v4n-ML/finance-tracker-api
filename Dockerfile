# --- Build Stage ---
# Use a specific Go version matching your development environment or project needs
FROM golang:1.24-alpine AS builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go module files first to leverage Docker cache
COPY go.mod go.sum ./
# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source code into the container
COPY . .

# Build the Go app
# - CGO_ENABLED=0 creates a static binary (needed for Alpine)
# - -ldflags="-w -s" reduces the size of the binary by removing debug information
# - -o builds the output binary named 'server' inside /app
# - The build target is your main package entry point
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/server

# --- Runtime Stage ---
# Use a minimal non-root image for security and size
FROM alpine:latest

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy only the built application binary from the build stage to the /app directory
COPY --from=builder /app/server /app/server

# EXPOSE port 8080. This is the port the Gin server inside the container will listen on.
# Make sure this matches the PORT environment variable your app uses (default 8080 seems correct for your setup)
EXPOSE 8080

# Command to run the executable
CMD ["/app/server"]