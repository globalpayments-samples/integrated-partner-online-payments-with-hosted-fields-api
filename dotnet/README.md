# .NET Card Payment Integration

This implementation demonstrates complete card payment processing using .NET Core with Global Payments hosted fields tokenization and JWT authentication.

## Requirements

- .NET 9.0 or later
- Global Payments account with JWT authentication enabled

## Project Structure

- `Program.cs` - ASP.NET Core minimal API with payment processing
- `.env.sample` - Environment variable template
- `run.sh` - Startup script
- `appsettings.json` - Application configuration
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
3. Restore dependencies:
   ```bash
   dotnet restore
   ```
4. Run the application:
   ```bash
   ./run.sh
   ```
   Or manually:
   ```bash
   dotnet run
   ```
5. Open [http://localhost:8000](http://localhost:8000) in your browser

## Implementation Details

### JWT Authentication
Uses System.IdentityModel.Tokens.Jwt for token generation:
- Creates JWT payload with account credential and region
- Signs token using HS256 algorithm with `AUTHTOKEN_JWT_SECRET`
- Token includes timestamp for validation
- Configured for AuthTokenV2 type

### ASP.NET Core Configuration
Minimal API implementation:
- Lightweight endpoint configuration
- Static file serving from parent directory
- Built-in JSON serialization
- Environment variable loading via dotenv.net
- Listens on configurable port (default: 8000)

### Payment Processing Flow
1. Client requests configuration via `/config` endpoint (receives `HOSTED_FIELDS_API_KEY`)
2. Hosted fields library initializes with API key
3. User enters card details in secure iframes
4. Hosted fields tokenize card data client-side
5. Client submits payment token, amount, and billing zip to `/process-payment`
6. Server constructs JWT token for API authentication
7. Server makes direct HTTP call to Global Payments endpoint using `TRANSACTIONS_API_KEY`
8. Payment is processed using HttpClient and transaction ID is returned

### Input Sanitization
- Postal codes: Regex validation, alphanumeric and hyphens only, max 10 characters
- Amounts: Decimal validation with positive value enforcement
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
- Secure error handling with proper HTTP status codes

## Production Considerations

- Use production hosting (Azure App Service, IIS, etc.)
- Enable HTTPS with proper certificates
- Configure rate limiting middleware
- Add comprehensive logging (Serilog, NLog)
- Implement health checks
- Use HttpClientFactory for connection pooling
- Configure CORS properly
- Add request validation middleware

---

## Troubleshooting

| Issue | Solution |
|-------|----------|
| `dotnet` command not found | Install .NET SDK 6.0+. Run `dotnet --version` to check |
| Build fails | Run `dotnet restore` before `dotnet run` |
| Port already in use | Set a different port in `launchSettings.json` or use `--urls` flag |
| `.env` not loading | Verify `.env` file exists in the language directory (not project root) |

---

## Resources

- [Parent Project README](../README.md)
- [Global Payments Developer Portal](https://developer.globalpayments.com/)
- [API Reference](https://developer.globalpayments.com/api/references-overview)
- [Test Cards](https://developer.globalpayments.com/resources/test-cards)
