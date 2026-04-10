# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o uptimer ./cmd/uptimer/main.go

# Final stage
FROM alpine:latest

# Install runtime dependencies
# ca-certificates are needed for outgoing HTTPS requests (monitoring)
# tzdata for correct time handling
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/uptimer .

# Ensure the binary is executable
RUN chmod +x ./uptimer

# Expose the port the app runs on
EXPOSE 8080

# Run the binary
CMD ["./uptimer"]
