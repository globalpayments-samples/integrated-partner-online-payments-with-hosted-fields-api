package com.globalpayments.example;

import io.github.cdimascio.dotenv.Dotenv;
import io.jsonwebtoken.Jwts;
import io.jsonwebtoken.SignatureAlgorithm;
import jakarta.servlet.ServletException;
import jakarta.servlet.annotation.WebServlet;
import jakarta.servlet.http.HttpServlet;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.OutputStream;
import java.math.BigDecimal;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;
import java.util.Date;
import java.util.HashMap;
import java.util.Map;
import java.util.UUID;

import javax.sound.midi.SysexMessage;

import org.json.JSONObject;

/**
 * Card Payment Processing Servlet
 *
 * This servlet demonstrates card payment processing using the Global Payments API.
 * It provides endpoints for configuration and payment processing, handling
 * tokenized card data to ensure secure payment processing.
 *
 * Endpoints:
 * - GET /config: Returns the public API key for client-side tokenization
 * - POST /process-payment: Processes card payments using tokenized data
 *
 * @author Global Payments
 * @version 1.0
 */

@WebServlet(urlPatterns = {"/process-payment", "/config"})
public class ProcessPaymentServlet extends HttpServlet {

    private static final long serialVersionUID = 1L;
    private final Dotenv dotenv = Dotenv.load();

    /**
     * Create JWT for authentication
     *
     * @return JWT token string
     */
    private String createJWT() {
        String key = dotenv.get("AUTHTOKEN_JWT_SECRET");

        Map<String, Object> claims = new HashMap<>();
        claims.put("type", "AuthTokenV2");
        claims.put("region", "US");
        claims.put("account_credential", dotenv.get("ACCOUNT_CREDENTIAL"));
        claims.put("ts", System.currentTimeMillis());

        return Jwts.builder()
            .setHeaderParam("typ", "JWT")
            .setClaims(claims)
            .signWith(SignatureAlgorithm.HS256, key.getBytes(StandardCharsets.UTF_8))
            .compact();
    }

    /**
     * Handles GET requests to /config endpoint.
     * Returns the public API key needed for client-side tokenization.
     *
     * @param request The HTTP request
     * @param response The HTTP response
     * @throws ServletException If there's an error in servlet processing
     * @throws IOException If there's an I/O error
     */
    @Override
    protected void doGet(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {
        if (request.getServletPath().equals("/config")) {
            response.setContentType("application/json");
            String apiKey = dotenv.get("HOSTED_FIELDS_API_KEY");

            JSONObject responseJson = new JSONObject();
            responseJson.put("success", true);

            JSONObject data = new JSONObject();
            data.put("apiKey", apiKey);
            responseJson.put("data", data);

            response.getWriter().write(responseJson.toString());
        }
    }

    /**
     * Sanitizes postal code input by removing invalid characters.
     * Only allows alphanumeric characters and hyphens, limited to 10 characters.
     *
     * @param postalCode The postal code to sanitize, can be null
     * @return A sanitized postal code containing only alphanumeric characters
     *         and hyphens, limited to 10 characters. Returns empty string if input is null.
     */
    private String sanitizePostalCode(String postalCode) {
        if (postalCode == null) {
            return "";
        }
        String sanitized = postalCode.replaceAll("[^a-zA-Z0-9-]", "");
        return sanitized.length() > 10 ? sanitized.substring(0, 10) : sanitized;
    }

    /**
     * Handles POST requests to /process-payment endpoint.
     * Processes card payments using tokenized card data.
     *
     * @param request The HTTP request containing payment details
     * @param response The HTTP response
     * @throws ServletException If there's an error in servlet processing
     * @throws IOException If there's an I/O error
     */
    @Override
    protected void doPost(HttpServletRequest request, HttpServletResponse response)
            throws ServletException, IOException {

        response.setContentType("application/json");

        try {
            // Validate and extract payment information
            String paymentToken = request.getParameter("payment_token");
            String billingZip = request.getParameter("billing_zip");
            String amountStr = request.getParameter("amount");

            if (paymentToken == null || billingZip == null || amountStr == null ||
                paymentToken.trim().isEmpty() || billingZip.trim().isEmpty() || amountStr.trim().isEmpty()) {
                throw new Exception("Missing required fields");
            }

            // Validate and parse amount
            BigDecimal amount;
            try {
                amount = new BigDecimal(amountStr);
                if (amount.compareTo(BigDecimal.ZERO) <= 0) {
                    throw new Exception("Amount must be a positive number");
                }
            } catch (NumberFormatException e) {
                throw new Exception("Invalid amount format");
            }

            String serviceUrl = "https://api.pit.paygateway.com";
            String endpoint = "/transactions/creditsales";

            // Build request body
            JSONObject requestBody = new JSONObject();
            requestBody.put("reference_id", UUID.randomUUID().toString());

            JSONObject card = new JSONObject();
            card.put("temporary_token", paymentToken);
            requestBody.put("card", card);

            JSONObject customer = new JSONObject();
            JSONObject billingAddress = new JSONObject();
            billingAddress.put("postal_code", billingZip);
            customer.put("billing_address", billingAddress);
            requestBody.put("customer", customer);

            JSONObject payment = new JSONObject();
            payment.put("amount", amountStr);
            payment.put("currency_code", "840");
            requestBody.put("payment", payment);

            JSONObject transaction = new JSONObject();
            transaction.put("country_code", "840");
            JSONObject processingIndicators = new JSONObject();
            processingIndicators.put("allow_duplicate", true);
            processingIndicators.put("create_token", true);
            processingIndicators.put("address_verification_service", true);
            transaction.put("processing_indicators", processingIndicators);
            requestBody.put("transaction", transaction);

            // Make HTTP request
            URL url = new URL(serviceUrl + endpoint);
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("POST");
            conn.setRequestProperty("Authorization", "AuthToken " + createJWT());
            conn.setRequestProperty("X-GP-Version", "2021-04-08");
            conn.setRequestProperty("X-GP-Api-Key", dotenv.get("TRANSACTIONS_API_KEY"));
            conn.setRequestProperty("X-GP-Partner-App-Name", "GP Integrated Hosted Fields Sample (Java)");
            conn.setRequestProperty("Content-Type", "application/json");
            conn.setDoOutput(true);

            try (OutputStream os = conn.getOutputStream()) {
                byte[] input = requestBody.toString().getBytes(StandardCharsets.UTF_8);
                os.write(input, 0, input.length);
            }

            int responseCode = conn.getResponseCode();
            StringBuilder responseBuilder = new StringBuilder();

            // Get the appropriate input stream based on response code
            java.io.InputStream inputStream = (responseCode >= 200 && responseCode < 300)
                ? conn.getInputStream()
                : conn.getErrorStream();

            // Handle case where stream might be null
            if (inputStream != null) {
                try (java.io.BufferedReader br = new java.io.BufferedReader(
                        new java.io.InputStreamReader(inputStream, StandardCharsets.UTF_8))) {
                    String responseLine;
                    while ((responseLine = br.readLine()) != null) {
                        responseBuilder.append(responseLine.trim());
                    }
                }
            } else {
                // If no response body available, create a generic error response
                responseBuilder.append("{\"status\":\"error\",\"message\":\"No response from server\"}");
            }

            JSONObject apiResponse = new JSONObject(responseBuilder.toString());

            // Verify transaction was successful
            if (!"approved".equals(apiResponse.optString("status"))) {
                response.setStatus(HttpServletResponse.SC_BAD_REQUEST);

                JSONObject errorResponse = new JSONObject();
                errorResponse.put("success", false);
                errorResponse.put("message", "Payment processing failed");

                JSONObject error = new JSONObject();
                error.put("code", "PAYMENT_DECLINED");
                error.put("details", apiResponse.optString("status"));
                errorResponse.put("error", error);

                response.getWriter().write(errorResponse.toString());
                return;
            }

            // Return success response with reference ID
            JSONObject successResponse = new JSONObject();
            successResponse.put("success", true);
            successResponse.put("message", "Payment successful! Reference ID: " + apiResponse.getString("reference_id"));

            JSONObject data = new JSONObject();
            data.put("reference_id", apiResponse.getString("reference_id"));
            successResponse.put("data", data);

            response.getWriter().write(successResponse.toString());

        } catch (Exception e) {
            // Handle payment processing errors
            response.setStatus(HttpServletResponse.SC_BAD_REQUEST);

            JSONObject errorResponse = new JSONObject();
            errorResponse.put("success", false);
            errorResponse.put("message", "Payment processing failed");

            JSONObject error = new JSONObject();
            error.put("code", "API_ERROR");
            error.put("details", e.getMessage());
            errorResponse.put("error", error);

            response.getWriter().write(errorResponse.toString());
        }
    }
}
