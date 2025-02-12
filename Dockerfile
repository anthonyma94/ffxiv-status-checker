# Build stage
FROM golang:1.23.4-alpine AS builder

WORKDIR /app
# Copy go.mod and go.sum first for caching dependencies.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code.
COPY . .

# Build the binary with CGO disabled for a fully static binary.
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ffxiv-status-checker .

# Final stage: create a minimal runtime image.
FROM alpine:latest

WORKDIR /app/
RUN mkdir data
VOLUME /app/data
# Copy the binary from the builder stage.
COPY --from=builder /app/ffxiv-status-checker .

CMD ["./ffxiv-status-checker"]
