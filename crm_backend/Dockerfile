# Build stage
FROM golang:alpine AS builder

WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build standard statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server cmd/server/main.go

# Final minimal stage
FROM alpine:3.19

WORKDIR /app

# Install standard ca-certificates and tzdata for remote requests and time management
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /app/bin/server .

EXPOSE 8080

CMD ["./server"]
