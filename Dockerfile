# --- Stage 1: The Builder ---
# This stage compiles the Go application
FROM golang:latest AS builder

# Set necessary environment variables
WORKDIR /app
ENV CGO_ENABLED=1
ENV GOOS=linux

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application into a single static binary.
# The -ldflags="-w -s" strips debug information, making the binary smaller.
RUN go build -ldflags="-w -s" -o /go-app .

# --- Stage 2: The Final Image ---
# This stage creates the tiny production image
FROM alpine:latest

# It's good practice to run as a non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

WORKDIR /app

# Copy the compiled binary from the 'builder' stage
COPY --from=builder /go-app .

# Create a directory for the SQLite database file. This is where we will mount our volume.
RUN mkdir /app/data

# Expose the port
EXPOSE 8000

# This is the command that will run when the container starts
CMD ["./go-app"]

# Linking db file to the container
# docker run -p 8000:8000 -v ./db-data:/app/data your-app-name