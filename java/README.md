# Java Card Payment Integration

This implementation demonstrates complete card payment processing using Java with Global Payments hosted fields tokenization and JWT authentication.

## Requirements

- Java 11 or later
- Maven
- Global Payments account with JWT authentication enabled

## Project Structure

- `src/main/java/com/globalpayments/example/` - Java servlet implementation
  - `ProcessPaymentServlet.java` - Payment processing endpoint
  - `ConfigServlet.java` - Configuration endpoint
- `src/main/webapp/WEB-INF/web.xml` - Web application configuration
- `.env.sample` - Environment variable template
- `pom.xml` - Maven dependencies (jakarta.servlet, java-jwt, gson, dotenv-java)
- `run.sh` - Startup script with Jetty server
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
3. Install dependencies and build:
   ```bash
   mvn clean install
   ```
4. Run the application:
   ```bash
   ./run.sh
   ```
   Or manually:
   ```bash
   mvn jetty:run
   ```
5. Open [http://localhost:8000](http://localhost:8000) in your browser

## Implementation Details

### JWT Authentication
Uses Auth0's Java-JWT library to generate authentication tokens:
- Creates JWT payload with account credential and region
- Signs token using HS256 algorithm with `AUTHTOKEN_JWT_SECRET`
- Token includes timestamp for validation
- Configured for AuthTokenV2 type

### Servlet Configuration
Jakarta EE servlet-based implementation:
- Separate servlets for configuration and payment processing
- JSON request/response handling with Gson
- Environment variable configuration via dotenv-java
- Runs on embedded Jetty server for development

### Payment Processing Flow
1. Client requests configuration via `/config` servlet (receives `HOSTED_FIELDS_API_KEY`)
2. Hosted fields library initializes with API key
3. User enters card details in secure iframes
4. Hosted fields tokenize card data client-side
5. Client submits payment token, amount, and billing zip to `/process-payment`
6. Server constructs JWT token for API authentication
7. Server makes direct HTTP call to Global Payments endpoint using `TRANSACTIONS_API_KEY`
8. Payment is processed and transaction ID is returned

### Input Sanitization
- Postal codes: Regex validation, alphanumeric and hyphens only, max 10 characters
- Amounts: Decimal validation, positive value enforcement
- Tokens: Presence and format validation

## API Endpoints

### GET /config
Returns hosted fields API key.

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

**Request/Response:** Same format as other implementations (see main README)

## Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `HOSTED_FIELDS_API_KEY` | API key for hosted fields client-side | Yes |
| `TRANSACTIONS_API_KEY` | API key for server-side transactions | Yes |
| `AUTHTOKEN_JWT_SECRET` | Secret key for JWT signing | Yes |
| `ACCOUNT_CREDENTIAL` | Global Payments account credential | Yes |

## Security Features

- PCI Compliance via hosted fields
- JWT Authentication
- Input sanitization and validation
- Environment variable credential storage
- Secure error handling

## Production Considerations

- Deploy to production servlet container (Tomcat, WildFly, etc.)
- Enable HTTPS
- Configure rate limiting
- Add comprehensive logging
- Implement proper exception handling
- Use connection pooling for HTTP requests

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `mvn` command not found | Install Maven 3.6+. Run `mvn -v` to check version |
| Build fails | Ensure Java 11+ is installed. Run `java -version` to check |
| Port already in use | Stop other services on port 8000, or modify `pom.xml` cargo config |
| `.env` not loading | Verify `.env` file exists in the language directory (not project root) |

---

## Resources

- [Parent Project README](../README.md)
- [Global Payments Developer Portal](https://developer.globalpayments.com/)
- [API Reference](https://developer.globalpayments.com/api/references-overview)
- [Test Cards](https://developer.globalpayments.com/resources/test-cards)
