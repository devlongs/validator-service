# validator-service

This service implements a backend API for validator creation and key management. Built with Golang and using the chi router, it leverages the repository pattern with dependency injection. The project is containerized using Docker.

## Features

- **POST /validators**: Create a validator request.
- **GET /validators/{request_id}**: Check the status of a validator request.
- **GET /health**: Health check endpoint.
- Asynchronous key generation simulation (20ms delay per key).
- Basic input validation (e.g., Ethereum address format).


## Running Locally

1. **Clone the repository:**

   ```bash
   git clone https://github.com/devlongs/validator-service.git
   cd validator-service
    ```

2. **Run the Application:**

   ```bash
   go run cmd/api/main.go
    ```
The server will start on port 8080.

3. **Run with Docker:**
Build the Docker image:

   ```bash
   docker build -t validator-service .
    ```
    Run the Docker container:

   ```bash
   docker run -p 8080:8080 validator-service
    ```


# API Documentation
## Create Validator Request
- Endpoint: POST /validators
- Request Body:
```json
{
  "num_validators": 5,
  "fee_recipient": "0x1234567890abcdef1234567890abcdef12345678"
}
```

- Response:
```json
{
  "request_id": "generated-uuid",
  "message": "Validator creation in progress"
}
```

## Check Validator Request Status
- Endpoint: GET /validators/{request_id}
- Response (if successful):
```json
{
  "status": "successful",
  "keys": [
    "key1",
    "key2",
    "key3"
  ]
}
```

- Response (if failed):
```json
{
  "status": "failed",
  "message": "Error processing request"
}
```

## Health Check
- Endpoint: GET /health
- Response:
```json
{
  "status": "healthy"
}
```

## Running Unit Tests
- Run tests with:
```bash
go test ./...
```
