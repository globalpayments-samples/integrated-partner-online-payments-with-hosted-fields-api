using dotenv.net;
using System.IdentityModel.Tokens.Jwt;
using System.Security.Claims;
using System.Text;
using System.Text.Json;
using Microsoft.IdentityModel.Tokens;

namespace CardPaymentSample;

/// <summary>
/// Card Payment Processing Application
///
/// This application demonstrates card payment processing using the Global Payments API.
/// It provides endpoints for configuration and payment processing, handling tokenized
/// card data to ensure secure payment processing.
/// </summary>
public class Program
{
    public static void Main(string[] args)
    {
        // Load environment variables from .env file
        DotEnv.Load();

        var builder = WebApplication.CreateBuilder(args);

        var app = builder.Build();

        // Configure static file serving for the payment form
        app.UseDefaultFiles();
        app.UseStaticFiles();

        ConfigureEndpoints(app);

        var port = System.Environment.GetEnvironmentVariable("PORT") ?? "8000";
        app.Urls.Add($"http://0.0.0.0:{port}");

        app.Run();
    }

    /// <summary>
    /// Creates a JWT for authentication
    /// </summary>
    /// <returns>JWT token string</returns>
    private static string CreateJWT()
    {
        var key = System.Environment.GetEnvironmentVariable("AUTHTOKEN_JWT_SECRET");
        var securityKey = new SymmetricSecurityKey(Encoding.UTF8.GetBytes(key));
        var credentials = new SigningCredentials(securityKey, SecurityAlgorithms.HmacSha256);

        var claims = new[]
        {
            new Claim("type", "AuthTokenV2"),
            new Claim("region", "US"),
            new Claim("account_credential", System.Environment.GetEnvironmentVariable("ACCOUNT_CREDENTIAL") ?? ""),
            new Claim("ts", DateTimeOffset.UtcNow.ToUnixTimeMilliseconds().ToString())
        };

        var token = new JwtSecurityToken(
            claims: claims,
            signingCredentials: credentials
        );

        return new JwtSecurityTokenHandler().WriteToken(token);
    }

    /// <summary>
    /// Configures the application's HTTP endpoints for payment processing.
    /// </summary>
    /// <param name="app">The web application to configure</param>
    private static void ConfigureEndpoints(WebApplication app)
    {
        // Configure HTTP endpoints
        app.MapGet("/config", () => Results.Ok(new
        {
            success = true,
            data = new
            {
                apiKey = System.Environment.GetEnvironmentVariable("HOSTED_FIELDS_API_KEY")
            }
        }));

        ConfigurePaymentEndpoint(app);
    }

    /// <summary>
    /// Sanitizes postal code input by removing invalid characters.
    /// </summary>
    /// <param name="postalCode">The postal code to sanitize. Can be null.</param>
    /// <returns>
    /// A sanitized postal code containing only alphanumeric characters and hyphens,
    /// limited to 10 characters. Returns empty string if input is null or empty.
    /// </returns>
    private static string SanitizePostalCode(string postalCode)
    {
        if (string.IsNullOrEmpty(postalCode)) return string.Empty;

        // Remove any characters that aren't alphanumeric or hyphen
        var sanitized = new string(postalCode.Where(c => char.IsLetterOrDigit(c) || c == '-').ToArray());

        // Limit length to 10 characters
        return sanitized.Length > 10 ? sanitized[..10] : sanitized;
    }

    /// <summary>
    /// Configures the payment processing endpoint that handles card transactions.
    /// </summary>
    /// <param name="app">The web application to configure</param>
    private static void ConfigurePaymentEndpoint(WebApplication app)
    {
        app.MapPost("/process-payment", async (HttpContext context) =>
        {
            // Parse form data from the request
            var form = await context.Request.ReadFormAsync();
            var billingZip = form["billing_zip"].ToString();
            var token = form["payment_token"].ToString();
            var amountStr = form["amount"].ToString();

            // Validate required fields are present
            if (string.IsNullOrEmpty(token) || string.IsNullOrEmpty(billingZip) || string.IsNullOrEmpty(amountStr))
            {
                return Results.BadRequest(new
                {
                    success = false,
                    message = "Payment processing failed",
                    error = new
                    {
                        code = "VALIDATION_ERROR",
                        details = "Missing required fields"
                    }
                });
            }

            // Validate and parse amount
            if (!decimal.TryParse(amountStr, out var amount) || amount <= 0)
            {
                return Results.BadRequest(new
                {
                    success = false,
                    message = "Payment processing failed",
                    error = new
                    {
                        code = "VALIDATION_ERROR",
                        details = "Amount must be a positive number"
                    }
                });
            }

            var serviceUrl = "https://api.pit.paygateway.com";
            var endpoint = "/transactions/creditsales";

            // try
            // {
                using var httpClient = new HttpClient();
                httpClient.DefaultRequestHeaders.Add("Authorization", "AuthToken " + CreateJWT());
                httpClient.DefaultRequestHeaders.Add("X-GP-Version", "2021-04-08");
                httpClient.DefaultRequestHeaders.Add("X-GP-Api-Key", System.Environment.GetEnvironmentVariable("TRANSACTIONS_API_KEY"));
                httpClient.DefaultRequestHeaders.Add("X-GP-Partner-App-Name", "GP Integrated Hosted Fields Sample (.NET)");

                var requestBody = new
                {
                    reference_id = Guid.NewGuid().ToString(),
                    card = new { temporary_token = token },
                    customer = new
                    {
                        billing_address = new
                        {
                            postal_code = billingZip
                        }
                    },
                    payment = new
                    {
                        amount = amountStr,
                        currency_code = "840"
                    },
                    transaction = new
                    {
                        country_code = "840",
                        processing_indicators = new
                        {
                            allow_duplicate = true,
                            create_token = true,
                            address_verification_service = true
                        }
                    }
                };

                var jsonContent = JsonSerializer.Serialize(requestBody);
                var content = new StringContent(jsonContent);
                content.Headers.Remove("Content-Type");
                content.Headers.Add("Content-Type", "application/json");

                var response = await httpClient.PostAsync(serviceUrl + endpoint, content);
                var responseContent = await response.Content.ReadAsStringAsync();
                var apiResponse = JsonSerializer.Deserialize<JsonElement>(responseContent);

                // Verify transaction was successful
                if (apiResponse.GetProperty("status").GetString() != "approved")
                {
                    return Results.BadRequest(new
                    {
                        success = false,
                        message = "Payment processing failed",
                        error = new
                        {
                            code = "PAYMENT_DECLINED",
                            details = apiResponse.GetProperty("status").GetString()
                        }
                    });
                }

                // Return success response with reference ID
                var referenceId = apiResponse.GetProperty("reference_id").GetString();
                return Results.Ok(new
                {
                    success = true,
                    message = $"Payment successful! Reference ID: {referenceId}",
                    data = new
                    {
                        reference_id = referenceId
                    }
                });
            // }
            // catch (Exception ex)
            // {
            //     // Handle payment processing errors
            //     return Results.BadRequest(new
            //     {
            //         success = false,
            //         message = "Payment processing failed",
            //         error = new
            //         {
            //             code = "API_ERROR",
            //             details = ex.Message
            //         }
            //     });
            // }
        });
    }
}
