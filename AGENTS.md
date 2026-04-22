# Integrated Partner Online Payments with Hosted Fields

> Process card payments via Global Payments hosted fields tokenization and JWT-authenticated direct API calls — implemented identically in Node.js, Python, PHP, Java, .NET, and Go.

## Critical Patterns

1. **Two separate API keys serve different roles.** `HOSTED_FIELDS_API_KEY` is sent to the client for initializing the hosted fields iframe (`GET /config`). `TRANSACTIONS_API_KEY` is used server-side only, passed as `X-GP-Api-Key` in the payment request header. Swapping them or exposing `TRANSACTIONS_API_KEY` to the client breaks the integration and leaks a secret.

2. **JWT is constructed locally — no SDK call.** All six implementations build the JWT manually using `HS256` with a payload of `{type: "AuthTokenV2", region: "US", account_credential, ts: <milliseconds>}`. The `Authorization` header value is `"AuthToken " + <jwt>` (not `"Bearer "`). Using a different prefix or algorithm will produce 401 errors.

3. **Go requires manual header casing.** Go's `net/http` automatically canonicalizes header keys (e.g., `X-Gp-Api-Key`). The Global Payments API requires exact casing (`X-GP-Api-Key`, `X-GP-Version`). The Go implementation works around this by writing to `req.Header[<key>]` directly as a map entry rather than using `req.Header.Set()` — see `go/main.go` lines ~272–276.

4. **`currency_code` and `country_code` use ISO 4217 numeric codes, not alpha codes.** The value `"840"` means USD/US, not `"USD"`. Using alpha codes causes the transaction to fail silently or be rejected.

## Repository Structure

### Node.js (Express)
- [`nodejs/server.js`](nodejs/server.js) — all logic: JWT creation (line ~34), `/config` and `/process-payment` handlers
- [`nodejs/index.html`](nodejs/index.html) — frontend: loads hosted fields library, calls `/config`, tokenizes card, submits form

### Python (Flask)
- [`python/server.py`](python/server.py) — all logic: JWT creation (line ~27), `/config` and `/process-payment` routes

### PHP
- [`php/process-payment.php`](php/process-payment.php) — payment handler: JWT creation, API call via curl helper
- [`php/config.php`](php/config.php) — config endpoint
- [`php/curl.php`](php/curl.php) — cURL HTTP client helper used by `process-payment.php`

### Java (Jakarta EE servlet)
- [`java/src/main/java/com/globalpayments/example/ProcessPaymentServlet.java`](java/src/main/java/com/globalpayments/example/ProcessPaymentServlet.java) — single servlet handling both `/config` (GET) and `/process-payment` (POST)

### .NET (ASP.NET Core minimal API)
- [`dotnet/Program.cs`](dotnet/Program.cs) — all logic: JWT creation, endpoint registration for `/config` and `/process-payment`

### Go (standard library)
- [`go/main.go`](go/main.go) — all logic: manual JWT construction (line ~90), `handleConfig` and `handlePayment` handlers; note header casing workaround (~line 200)

### Shared
- [`nodejs/index.html`](nodejs/index.html) — canonical frontend, symlinked or copied into each language's static directory
- `docker-compose.yml` — runs all implementations in parallel for comparison
- `.env.sample` in each language directory — required credential template

## API Surface

| Method | Path | Purpose |
|--------|------|---------|
| GET | `/config` | Returns `HOSTED_FIELDS_API_KEY` to client for hosted fields init |
| POST | `/process-payment` | Accepts form-encoded `payment_token`, `billing_zip`, `amount`; calls Global Payments `/transactions/creditsales`; returns `reference_id` |

All six implementations expose identical endpoints. All implementations default to port 8000.

## Environment Variables

```bash
HOSTED_FIELDS_API_KEY=   # sent to browser — initializes the hosted fields iframe
TRANSACTIONS_API_KEY=    # server-side only — authenticates payment API requests
AUTHTOKEN_JWT_SECRET=    # signs the JWT; never leave the server
ACCOUNT_CREDENTIAL=      # included in JWT payload as account_credential
PORT=8000                # optional; defaults to 8000 if unset
```

Copy `.env.sample` → `.env` in the language directory before running.

## Payment Request Structure

The body sent to `https://api.pit.paygateway.com/transactions/creditsales`:

- `card.temporary_token` — the token produced by hosted fields (not a raw card number)
- `payment.currency_code: "840"` — numeric ISO 4217 for USD
- `transaction.country_code: "840"` — numeric ISO 3166 for US
- `transaction.processing_indicators.create_token: true` — vaults the card for reuse
- `transaction.processing_indicators.address_verification_service: true` — triggers AVS on the zip

## Architecture Summary

**Tokenization flow:** Browser loads hosted fields JS from CDN → fetches API key via `/config` → renders card iframe → user enters card → library tokenizes card client-side → `payment_token` returned to browser

**Payment flow:** Browser submits form with `payment_token` + `billing_zip` + `amount` → server builds JWT + request body → POST to `api.pit.paygateway.com/transactions/creditsales` → checks `status === "approved"` → returns `reference_id`

## Security Notes

No card data passes through the server — the hosted fields library handles card entry in an isolated iframe and returns only a `temporary_token`. The demo runs over HTTP; production deployments require HTTPS. The `/config` endpoint exposes `HOSTED_FIELDS_API_KEY` publicly by design — it is not a secret — but `TRANSACTIONS_API_KEY` and `AUTHTOKEN_JWT_SECRET` must never reach the client.

## SDK Versions

- Node.js: `jsonwebtoken` ^9.0, `express` ^4.18
- Python: `PyJWT`, `flask`, `requests`
- PHP: `firebase/php-jwt`, `vlucas/phpdotenv`
- Java: `io.jsonwebtoken:jjwt`, Jakarta EE servlet container
- .NET: `System.IdentityModel.Tokens.Jwt`, ASP.NET Core minimal API
- Go: `github.com/golang-jwt/jwt` not used — JWT is constructed manually with `crypto/hmac`
