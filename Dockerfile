# Use official Golang image as the builder
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy the entire workspace
COPY . .

# Build the settler binary
# Note: We build from the cmd/settler directory which uses go.work
RUN go build -o settler ./cmd/settler

# Use a minimal alpine image for the final runtime
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from the builder
COPY --from=builder /app/settler .

# Default command
ENTRYPOINT ["./settler"]
CMD ["help"]
