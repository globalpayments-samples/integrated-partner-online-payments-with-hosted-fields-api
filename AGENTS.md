# Integrated Partner Online Payments with Hosted Fields

> Process credit card payments using Global Payments hosted fields tokenization, server-side JWT authentication, and direct REST API calls — implemented identically in Node.js, Python, Go, .NET, Java, and PHP.

## Critical Patterns

1. **`Authorization` header uses `AuthToken` prefix, not `Bearer`.** The GP payments API authenticates with a custom scheme: `Authorization: AuthToken <JWT>`. Using `Bearer` will result in a 401 that is easy to misread. See `createJWT()` / `create_jwt()` / `CreateJWT()` in every server file.

2. **`currency_code` and `country_code` are numeric ISO codes, not alpha codes.** Both fields must be `"840"` (ISO-4217/ISO-3166-1 numeric for USD/US). Passing `"USD"` or `"US"` will produce a validation error from the API. See the `payment` and `transaction` objects in every server implementation.

3. **Go's `X-GP-*` headers must be set via map assignment, not `Header.Set()`.** Go's HTTP client auto-canonicalizes header names — `Header.Set("X-GP-Api-Key", ...)` silently becomes `X-Gp-Api-Key`, which the API rejects. The Go implementation bypasses this with direct map assignment: `req.Header["X-GP-Api-Key"] = []string{...}`. See `handlePayment()` in [`go/main.go`](go/main.go) (line ~272).

4. **PHP's `composer.json` lists `globalpayments/php-sdk` but it is never used.** The PHP implementation calls the GP REST API directly via `curl.php` using `firebase/php-jwt` for token signing. The SDK dependency is an artifact — do not build on it. See [`php/process-payment.php`](php/process-payment.php) and [`php/curl.php`](php/curl.php).

## Repository Structure

### Node.js (Express)
- [`nodejs/server.js`](nodejs/server.js) — single-file server; `createJWT()` builds the auth token, `app.get('/config')` returns the hosted fields key, `app.post('/process-payment')` calls the GP API
- [`nodejs/index.html`](nodejs/index.html) — payment form with hosted fields integration
- [`nodejs/.env.sample`](nodejs/.env.sample) — environment variable template

### Python (Flask)
- [`python/server.py`](python/server.py) — single-file server; `create_jwt()`, `get_config()`, `process_payment()`
- [`python/index.html`](python/index.html) — payment form
- [`python/.env.sample`](python/.env.sample) — environment variable template

### Go (standard library)
- [`go/main.go`](go/main.go) — single-file server; `createJWT()`, `handleConfig()`, `handlePayment()`; static files served from `static/`
- [`go/.env.sample`](go/.env.sample) — environment variable template

### .NET (ASP.NET Core / net9.0)
- [`dotnet/Program.cs`](dotnet/Program.cs) — single-file server; `CreateJWT()`, `ConfigureEndpoints()`, `ConfigurePaymentEndpoint()`; static files from `wwwroot/`
- [`dotnet/.env.sample`](dotnet/.env.sample) — environment variable template

### Java (Jakarta EE / Tomcat 10 via Cargo)
- [`java/src/main/java/com/globalpayments/example/ProcessPaymentServlet.java`](java/src/main/java/com/globalpayments/example/ProcessPaymentServlet.java) — single servlet handling both `/config` (GET via `doGet()`) and `/process-payment` (POST via `doPost()`); `createJWT()` for auth
- [`java/src/main/webapp/index.html`](java/src/main/webapp/index.html) — payment form
- [`java/.env.sample`](java/.env.sample) — environment variable template

### PHP (built-in server)
- [`php/process-payment.php`](php/process-payment.php) — payment handler; `createJWT()`, `sanitizePostalCode()`; delegates HTTP call to `curl.php`
- [`php/config.php`](php/config.php) — config endpoint returning `HOSTED_FIELDS_API_KEY`
- [`php/curl.php`](php/curl.php) — reusable cURL wrapper (note: SSL verification is disabled — not safe for production)
- [`php/.env.sample`](php/.env.sample) — environment variable template

### Shared
- [`docker-compose.yml`](docker-compose.yml) — multi-language Docker Compose configuration
- [`Dockerfile.tests`](Dockerfile.tests) — test runner image

## API Surface

All language implementations expose identical endpoints:

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/config` | Returns `HOSTED_FIELDS_API_KEY` for client-side hosted fields initialization |
| POST | `/process-payment` | Accepts `payment_token`, `billing_zip`, `amount`; proxies to GP API; returns `reference_id` |

## Environment Variables

All six language directories contain an identical `.env.sample`:

```bash
HOSTED_FIELDS_API_KEY=       # Public key passed to the hosted fields JS library (client-side)
TRANSACTIONS_API_KEY=        # Server-side key sent as X-GP-Api-Key on each payment request
AUTHTOKEN_JWT_SECRET=        # HS256 signing secret for JWT generation
ACCOUNT_CREDENTIAL=          # GP account credential embedded in the JWT payload
PORT=                        # Optional; all implementations default to 8000
```

## Sandbox Credentials

The `.env.sample` files contain working sandbox keys for the GP PIT (payment integration test) environment. The API base URL is `https://api.pit.paygateway.com`. Get production credentials at [developer.globalpayments.com](https://developer.globalpayments.com/).

## API Request Shape

All implementations POST to `https://api.pit.paygateway.com/transactions/creditsales` with:

**Required headers:**
- `Authorization: AuthToken <JWT>` — custom auth scheme; `AuthToken` prefix is exact
- `X-GP-Version: 2021-04-08` — API version; must match exactly
- `X-GP-Api-Key: <TRANSACTIONS_API_KEY>` — server-side API key
- `Content-Type: application/json`

**Non-obvious body fields:**
- `payment.currency_code` — `"840"` (numeric ISO-4217 for USD, not `"USD"`)
- `transaction.country_code` — `"840"` (numeric ISO-3166-1 for US, not `"US"`)
- `card.temporary_token` — the token produced by the hosted fields library, not raw card data
- `transaction.processing_indicators.create_token` — `true` vaults the card for reuse

## Architecture Summary

**Tokenization flow:** Browser loads hosted fields JS → user enters card in isolated iframes → library tokenizes card client-side → `payment_token` sent to server (no raw card data ever touches the server)

**Payment flow:** Server receives `payment_token` + `billing_zip` + `amount` → builds JWT from `AUTHTOKEN_JWT_SECRET` + `ACCOUNT_CREDENTIAL` → POSTs to GP API `/transactions/creditsales` → returns `reference_id` to browser on `status: approved`

## Security Notes

`php/curl.php` disables SSL peer and host verification (`CURLOPT_SSL_VERIFYPEER => false`). Neither `/config` nor `/process-payment` enforce caller authentication — any client can retrieve the hosted fields key or trigger a payment. These are demo shortcuts and must be addressed before production use.

## SDK Versions

- **Node.js**: `express` ^4.18.2, `jsonwebtoken` ^9.0.2
- **Python**: `Flask` 3.0.0, `PyJWT` 2.8.0, `requests` 2.31.0
- **Go**: `github.com/google/uuid` v1.5.0, `github.com/joho/godotenv` v1.5.1 (no GP SDK)
- **.NET**: `DotEnv.Net` 3.2.1, `System.IdentityModel.Tokens.Jwt` 7.0.3 (net9.0)
- **Java**: `jjwt` 0.9.1, `dotenv-java` 3.0.0, Jakarta Servlet 5.0 / Tomcat 10.x
- **PHP**: `firebase/php-jwt` ^6.11, `vlucas/phpdotenv` ^5.5 (GP PHP SDK listed in `composer.json` but unused)
