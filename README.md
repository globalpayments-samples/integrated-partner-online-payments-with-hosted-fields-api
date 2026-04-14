# Integrated Partner Online Payments with Hosted Fields

This project demonstrates complete card payment processing using Global Payments hosted fields tokenization across 6 programming languages. Each implementation provides a fully functional payment integration with JWT authentication and direct API communication.

## Available Implementations

- [.NET Core](./dotnet/) - ASP.NET Core web application
- [Go](./go/) - Go HTTP server application
- [Java](./java/) - Jakarta EE servlet-based web application
- [Node.js](./nodejs/) - Express.js web application
- [PHP](./php/) - PHP web application
- [Python](./python/) - Flask web application

## Key Features

- **JWT Authentication** - Secure token-based authentication with Global Payments API
- **Hosted Fields Tokenization** - PCI-compliant card data capture using client-side hosted fields
- **Direct API Integration** - Server-to-server payment processing via REST API
- **Cross-language Consistency** - Identical functionality across all 6 language implementations

## How It Works

### Authentication Flow
1. Server generates JWT using `AUTHTOKEN_JWT_SECRET` and `ACCOUNT_CREDENTIAL`
2. Client initializes hosted fields with the JWT token
3. User enters card details in secure, isolated iframes
4. Hosted fields library tokenizes card data client-side

### Payment Processing Flow
1. Client submits tokenized card data and billing zip code
2. Server receives payment token and billing information
3. Server makes direct API call to Global Payments endpoint
4. Payment is processed and transaction ID is returned
5. Results are displayed to the user

### Architecture
- **Client-side**: Global Payments hosted fields library handles secure card data entry
- **Server-side**: JWT creation, API request construction, and payment processing
- **API**: Direct REST communication with Global Payments payment endpoints

## Quick Start

1. **Choose your language** - Navigate to any implementation directory:
   - [nodejs](./nodejs/) - Node.js with Express
   - [python](./python/) - Python with Flask
   - [php](./php/) - PHP with built-in server
   - [java](./java/) - Java with Jakarta EE
   - [dotnet](./dotnet/) - .NET Core
   - [go](./go/) - Go with standard library

2. **Configure credentials** - Copy `.env.sample` to `.env` and add your credentials:
   ```bash
   HOSTED_FIELDS_API_KEY=your_hosted_fields_api_key
   TRANSACTIONS_API_KEY=your_transactions_api_key
   AUTHTOKEN_JWT_SECRET=your_jwt_secret
   ACCOUNT_CREDENTIAL=your_account_credential
   ```

3. **Run the server** - Execute the run script:
   ```bash
   ./run.sh
   ```
   The server will start on `http://localhost:8000` (or port 8888 for Go)

4. **Test the integration** - Open your browser and complete a test payment

## Environment Variables

Each implementation requires these environment variables:

- `HOSTED_FIELDS_API_KEY` - API key for hosted fields client-side initialization
- `TRANSACTIONS_API_KEY` - API key for server-side transaction processing
- `AUTHTOKEN_JWT_SECRET` - Secret key for JWT signing
- `ACCOUNT_CREDENTIAL` - Your Global Payments account credential
- `PORT` (optional) - Server port (defaults to 8000, or 8888 for Go)

## API Endpoints

All implementations provide identical API endpoints:

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
  "message": "Payment declined: Insufficient funds"
}
```

## Prerequisites

- **Global Payments Account** with JWT authentication enabled
- **Development Environment** for your chosen language:
  - Node.js 14+ (for Node.js implementation)
  - Python 3.7+ (for Python implementation)
  - PHP 7.4+ (for PHP implementation)
  - Java 11+ (for Java implementation)
  - .NET 6.0+ (for .NET implementation)
  - Go 1.23+ (for Go implementation)
- **Package Manager** (npm, pip, composer, maven, dotnet, or go mod)

## Security Considerations

This example demonstrates production-ready security patterns:

- **PCI Compliance** - No card data touches your server (handled by hosted fields)
- **JWT Authentication** - Secure, time-limited tokens for API access
- **Input Sanitization** - Postal codes and amounts are validated and sanitized
- **HTTPS Required** - Always use HTTPS in production environments
- **Environment Variables** - Sensitive credentials stored outside source code
- **Error Handling** - Secure error messages that don't leak sensitive information

## Resources

- [Global Payments Developer Portal](https://developer.globalpayments.com/)
- [API Reference](https://developer.globalpayments.com/api/references-overview)

## Community

- 🌐 **Developer Portal** — [developer.globalpayments.com](https://developer.globalpayments.com)
- 💬 **Discord** — [Join the community](https://discord.gg/myER9G9qkc)
- 📋 **GitHub Discussions** — [github.com/globalpayments-samples](https://github.com/globalpayments-samples)
- 📧 **Newsletter** — [Subscribe](https://www.globalpayments.com/en-gb/modals/newsletter)
- 💼 **LinkedIn** — [Global Payments for Developers](https://www.linkedin.com/showcase/global-payments-for-developers/posts/?feedView=all)

Have a question or found a bug? [Open an issue](https://github.com/globalpayments-samples/integrated-partner-online-payments-with-hosted-fields-api/issues) or reach out at [communityexperience@globalpay.com](mailto:communityexperience@globalpay.com).

## License

MIT
