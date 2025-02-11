# Build stage
FROM golang:1.18 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o validator-service ./cmd/api

# Final stage
FROM debian:bullseye-slim
COPY --from=builder /app/validator-service /validator-service
EXPOSE 8080
ENTRYPOINT ["/validator-service"]
