# Go Card Payment Integration

This implementation demonstrates complete card payment processing using Go with Global Payments hosted fields tokenization and JWT authentication.

## Requirements

- Go 1.23 or later
- Global Payments account with JWT authentication enabled

## Project Structure

- `main.go` - HTTP server with JWT creation and payment processing
- `go.mod` - Go module dependencies (jwt-go, godotenv)
- `.env.sample` - Environment variable template
- `run.sh` - Startup script
- `../index.html` - Shared client-side payment form (parent directory)

## Setup

1. Copy `.env.sample` to `.env`
2. Update `.env` with your Global Payments credentials:
   ```bash
   HOSTED_FIELDS_API_KEY=your_hosted_fields_api_key
   TRANSACTIONS_API_KEY=your_transactions_api_key
   AUTHTOKEN_JWT_SECRET=your_jwt_secret
   ACCOUNT_CREDENTIAL=your_account_credential
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Run the application:
   ```bash
   ./run.sh
   ```
   Or manually:
   ```bash
   go run main.go
   ```
5. Open [http://localhost:8888](http://localhost:8888) in your browser

## Implementation Details

### JWT Authentication
Uses golang-jwt/jwt library for token generation:
- Creates JWT payload with account credential and region
- Signs token using HS256 algorithm with `AUTHTOKEN_JWT_SECRET`
- Token includes timestamp for validation
- Configured for AuthTokenV2 type

### Server Configuration
Standard library HTTP server:
- Native http.ServeMux for routing
- Static file serving from parent directory
- JSON encoding/decoding with encoding/json
- Environment variable loading via godotenv
- Listens on port 8888 by default

### Payment Processing Flow
1. Client requests configuration via `/config` endpoint (receives `HOSTED_FIELDS_API_KEY`)
2. Hosted fields library initializes with API key
3. User enters card details in secure iframes
4. Hosted fields tokenize card data client-side
5. Client submits payment token, amount, and billing zip to `/process-payment`
6. Server constructs JWT token for API authentication
7. Server makes direct HTTP call to Global Payments endpoint using `TRANSACTIONS_API_KEY`
8. Payment is processed using http.Client and transaction ID is returned

### Input Sanitization
- Postal codes: Regex validation, alphanumeric and hyphens only, max 10 characters
- Amounts: Float parsing with positive value validation
- Tokens: Presence and format validation
- Uses regexp package for validation

## API Endpoints

### GET /config
Returns hosted fields API key.

**Response:**
```json
{
  "success": true,
  "data": {
    "apiKey": "ALALaW0WQoKZ8MjFNfGm7J7FA6UFYCc4"
  }
}
```

### POST /process-payment
Processes a card payment using tokenized card data.

**Request/Response:** Same format as other implementations (see main README)

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `HOSTED_FIELDS_API_KEY` | API key for hosted fields client-side | Yes |
| `TRANSACTIONS_API_KEY` | API key for server-side transactions | Yes |
| `AUTHTOKEN_JWT_SECRET` | Secret key for JWT signing | Yes |
| `ACCOUNT_CREDENTIAL` | Global Payments account credential | Yes |

## Testing

Use test card numbers from Global Payments (see main README for details).

## Security Features

- PCI Compliance via hosted fields
- JWT Authentication
- Input sanitization and validation
- Environment variable credential storage
- Secure error handling with proper HTTP status codes

## Production Considerations

- Use reverse proxy (Nginx, Caddy)
- Enable HTTPS with TLS certificates
- Implement rate limiting middleware
- Add structured logging (logrus, zap)
- Use context for request timeouts
- Implement graceful shutdown
- Configure CORS properly
- Add request ID tracing
- Use connection pooling for HTTP client

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `go` command not found | Install Go 1.21+. Run `go version` to check |
| Module download fails | Run `go mod tidy` to resolve dependencies |
| Port already in use | Change port in source or set `PORT` environment variable |
| `.env` not loading | Verify `.env` file exists in the language directory (not project root) |

---

## Resources

- [Parent Project README](../README.md)
- [Global Payments Developer Portal](https://developer.globalpayments.com/)
- [API Reference](https://developer.globalpayments.com/api/references-overview)
- [Test Cards](https://developer.globalpayments.com/resources/test-cards)
