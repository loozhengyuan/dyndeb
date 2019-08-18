# Create builder image
FROM golang:1 AS builder

# Set working directoy
WORKDIR /app

# Install go modules
COPY go.mod go.sum ./
RUN go mod download

# Verify downloaded modules with go.sum
RUN go mod verify

# Build minary
COPY main.go preseed.cfg ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o dyndeb .

# Create production image
FROM alpine:latest

# Set working directory
WORKDIR /app

# Enable SSL/TLS
RUN apk --no-cache add ca-certificates

# Copy binary
COPY --from=builder /app .

# Set entrypoint
ENTRYPOINT ["./dyndeb"]