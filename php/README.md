# PHP Card Payment Integration

This implementation demonstrates complete card payment processing using PHP with Global Payments hosted fields tokenization and JWT authentication.

## Requirements

- PHP 7.4 or later (PHP 8.2+ recommended)
- Composer
- Global Payments account with JWT authentication enabled

## Project Structure

- `config.php` - Configuration and JWT generation
- `process-payment.php` - Payment processing endpoint
- `composer.json` - Project dependencies (firebase/php-jwt, vlucas/phpdotenv, guzzlehttp/guzzle)
- `.env.sample` - Environment variable template
- `run.sh` - Startup script with built-in PHP server
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
4. Install dependencies:
   ```bash
   composer install
   ```
5. Run the application:
   ```bash
   ./run.sh
   ```
   Or manually:
   ```bash
   php -S localhost:8000
   ```

5. Open [http://localhost:8000](http://localhost:8000) in your browser

## Implementation Details

### JWT Authentication
The server generates JWT tokens for hosted fields authentication:
- Creates JWT payload with account credential and region
- Signs token using HS256 algorithm with `AUTHTOKEN_JWT_SECRET`
- Token includes timestamp for validation
- Uses Firebase PHP-JWT library

### Application Structure
PHP-based implementation:
- Uses Composer for dependency management
- Built-in PHP server for development
- Separate endpoints for config and payment processing
- JSON responses with proper headers

### Payment Processing Flow
1. Client requests configuration via `/config` endpoint (receives `HOSTED_FIELDS_API_KEY`)
2. Hosted fields library initializes with API key
3. User enters card details in secure iframes
4. Hosted fields tokenize card data client-side
5. Client submits payment token, amount, and billing zip to `/process-payment`
6. Server constructs JWT token for API authentication
7. Server makes direct API call to Global Payments endpoint using `TRANSACTIONS_API_KEY`
8. Payment is processed using Guzzle HTTP client
9. Transaction ID is extracted and returned to client

### Input Sanitization
Implements robust input validation:
- **Postal codes**: Regex validation, removes non-alphanumeric characters (except hyphens)
- **Amounts**: Parses and validates as decimal, ensures positive value
- **Tokens**: Validates presence and format
- Prevents injection attacks and malformed data

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

- **PCI Compliance** - Card data never touches your server
- **JWT Authentication** - Secure tokens for API access
- **Input Sanitization** - All inputs validated and sanitized
- **Environment Variables** - Credentials stored securely
- **Error Handling** - Generic error messages
- **CORS Headers** - Properly configured

## Production Considerations

- Use Apache/Nginx instead of built-in PHP server
- Enable HTTPS (required for PCI compliance)
- Configure rate limiting
- Implement CSRF protection
- Set proper PHP security directives (disable_functions, open_basedir)
- Add comprehensive logging
- Implement idempotency keys
