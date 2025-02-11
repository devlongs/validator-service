# Build stage
FROM golang:1.18 as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -o validator-service ./cmd/server

# Final stage
FROM scratch
COPY --from=builder /app/validator-service /validator-service
EXPOSE 8080
ENTRYPOINT ["/validator-service"]
