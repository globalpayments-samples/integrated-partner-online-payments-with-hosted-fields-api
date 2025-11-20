# Python Card Payment Integration

This implementation demonstrates complete card payment processing using Python/Flask with Global Payments hosted fields tokenization and JWT authentication.

## Requirements

- Python 3.7 or later
- pip (Python Package Installer)
- Global Payments account with JWT authentication enabled

## Project Structure

- `server.py` - Flask server with JWT creation and payment processing
- `requirements.txt` - Project dependencies (Flask, python-dotenv, PyJWT, requests)
- `.env.sample` - Environment variable template
- `run.sh` - Startup script with virtual environment creation
- `../index.html` - Shared client-side payment form (parent directory)

## Setup

1. Copy `.env.sample` to `.env`
2. Update `.env` with your Global Payments credentials:
   ```bash
   HOSTED_FIELDS_API_KEY=your_hosted_fields_api_key
   TRANSACTIONS_API_KEY=your_transactions_api_key
   AUTHTOKEN_JWT_SECRET=your_jwt_secret
   ACCOUNT_CREDENTIAL=your_account_credential
   PORT=8000
   ```
3. Create and activate a virtual environment (recommended):
   ```bash
   python -m venv venv
   source venv/bin/activate  # On Windows: venv\Scripts\activate
   ```
4. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```
5. Run the application:
   ```bash
   ./run.sh
   ```
   Or manually:
   ```bash
   python server.py
   ```
6. Open [http://localhost:8000](http://localhost:8000) in your browser

## Implementation Details

### JWT Authentication
The server generates JWT tokens for hosted fields authentication:
- Creates JWT payload with account credential and region
- Signs token using HS256 algorithm with `AUTHTOKEN_JWT_SECRET`
- Token includes timestamp (milliseconds since epoch) for validation
- Configures token for AuthTokenV2 type

### Server Configuration
Flask server setup:
- Serves static files from current directory
- Parses JSON request bodies
- Returns JSON responses with appropriate status codes
- Listens on configurable port (default: 8000)

### Payment Processing Flow
1. Client requests configuration via `/config` endpoint (receives `HOSTED_FIELDS_API_KEY`)
2. Hosted fields library initializes with API key
3. User enters card details in secure iframes
4. Hosted fields tokenize card data client-side
5. Client submits payment token, amount, and billing zip to `/process-payment`
6. Server constructs JWT token for API authentication
7. Server makes direct API call to Global Payments endpoint using `TRANSACTIONS_API_KEY`
8. Payment is processed and transaction ID is returned
9. Results are displayed to the client

### Input Sanitization
Implements robust input validation:
- **Postal codes**: Removes non-alphanumeric characters (except hyphens), limits to 10 characters
- **Amounts**: Parses and validates as decimal, ensures positive value
- **Tokens**: Validates presence and format
- Prevents injection attacks and malformed data

### Direct API Integration
Server-to-server payment processing:
- Constructs REST API request with transaction details
- Includes JWT authentication header (`Authorization: AuthToken <jwt>`)
- Includes API key header (`X-GP-Api-Key`)
- Sends JSON payload with payment token and billing data
- Handles API responses and errors with proper status codes
- Extracts transaction IDs from successful payments

## API Endpoints

### GET /config
Returns hosted fields API key for client-side initialization.

**Response:**
```json
{
  "success": true,
  "data": {
    "apiKey": "your_hosted_fields_api_key"
  }
}
```

### POST /process-payment
Processes a card payment using tokenized card data.

**Request:**
```json
{
  "payment_token": "PMT_xxxxx",
  "billing_zip": "12345",
  "amount": "10.00"
}
```

**Response (Success):**
```json
{
  "success": true,
  "message": "Payment successful! Transaction ID: TRN_xxxxx"
}
```

**Response (Error):**
```json
{
  "success": false,
  "message": "Payment processing error: [error details]"
}
```

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `HOSTED_FIELDS_API_KEY` | API key for hosted fields client-side | Yes |
| `TRANSACTIONS_API_KEY` | API key for server-side transactions | Yes |
| `AUTHTOKEN_JWT_SECRET` | Secret key for JWT signing | Yes |
| `ACCOUNT_CREDENTIAL` | Global Payments account credential | Yes |
| `PORT` | Server port (default: 8000) | No |

## Security Features

This implementation includes production-ready security:

- **PCI Compliance** - Card data never touches your server (hosted fields handle tokenization)
- **JWT Authentication** - Secure, time-limited tokens for API access
- **Input Sanitization** - All user inputs are validated and sanitized
- **Environment Variables** - Credentials stored securely outside source code
- **Error Handling** - Generic error messages prevent information disclosure
- **JSON Responses** - Consistent API response format

## Production Considerations

Before deploying to production:
- Use a production-grade WSGI server (Gunicorn, uWSGI)
- Enable HTTPS (required for PCI compliance)
- Configure rate limiting to prevent abuse
- Add comprehensive logging and monitoring
- Implement CSRF protection
- Implement idempotency keys for payment retries
- Set up webhook handling for async payment notifications
- Add request timeout handling
- Implement proper error tracking
