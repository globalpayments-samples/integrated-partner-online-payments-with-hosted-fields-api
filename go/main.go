// Package main implements a card payment processing server using the Global Payments API.
// It provides endpoints for configuration and payment processing, handling tokenized
// card data to ensure secure payment processing.
package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

// Config represents the configuration response sent to the client
type Config struct {
	ApiKey string `json:"apiKey"`
}

// Response represents a standardized API response
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo represents error details in the response
type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

// PaymentAPIRequest represents the payment request to Global Payments API
type PaymentAPIRequest struct {
	ReferenceID string                `json:"reference_id"`
	Card        CardToken             `json:"card"`
	Customer    Customer              `json:"customer"`
	Payment     Payment               `json:"payment"`
	Transaction TransactionIndicators `json:"transaction"`
}

// CardToken represents the tokenized card data
type CardToken struct {
	TemporaryToken string `json:"temporary_token"`
}

// Customer represents customer billing information
type Customer struct {
	BillingAddress BillingAddress `json:"billing_address"`
}

// BillingAddress represents the billing address
type BillingAddress struct {
	PostalCode string `json:"postal_code"`
}

// Payment represents payment amount and currency
type Payment struct {
	Amount       string `json:"amount"`
	CurrencyCode string `json:"currency_code"`
}

// TransactionIndicators represents transaction processing indicators
type TransactionIndicators struct {
	CountryCode          string               `json:"country_code"`
	ProcessingIndicators ProcessingIndicators `json:"processing_indicators"`
}

// ProcessingIndicators represents payment processing flags
type ProcessingIndicators struct {
	AllowDuplicate             bool `json:"allow_duplicate"`
	CreateToken                bool `json:"create_token"`
	AddressVerificationService bool `json:"address_verification_service"`
}

// PaymentAPIResponse represents the API response from Global Payments
type PaymentAPIResponse struct {
	Status      string `json:"status"`
	ReferenceID string `json:"reference_id"`
}

// createJWT creates a JWT for authentication
func createJWT() string {
	secret := os.Getenv("AUTHTOKEN_JWT_SECRET")
	accountCredential := os.Getenv("ACCOUNT_CREDENTIAL")

	// Create payload
	payload := map[string]interface{}{
		"type":               "AuthTokenV2",
		"region":             "US",
		"account_credential": accountCredential,
		"ts":                 time.Now().UnixMilli(),
	}

	// Create header
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}

	// Encode header and payload
	headerJSON, _ := json.Marshal(header)
	payloadJSON, _ := json.Marshal(payload)

	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create signature
	message := headerB64 + "." + payloadB64
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	signature := base64.RawURLEncoding.EncodeToString(h.Sum(nil))

	return message + "." + signature
}

// sanitizePostalCode removes invalid characters from the postal code input.
// It only allows alphanumeric characters and hyphens, limiting the length to 10 characters.
// This handles both US (12345, 12345-6789) and international postal codes.
func sanitizePostalCode(postalCode string) string {
	// Remove any characters that aren't alphanumeric or hyphen
	reg := regexp.MustCompile("[^a-zA-Z0-9-]")
	sanitized := reg.ReplaceAllString(postalCode, "")
	// Limit length to 10 characters
	if len(sanitized) > 10 {
		return sanitized[:10]
	}
	return sanitized
}

// handleConfig handles the /config endpoint
func handleConfig(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	response := Response{
		Success: true,
		Data: Config{
			ApiKey: os.Getenv("HOSTED_FIELDS_API_KEY"),
		},
	}
	json.NewEncoder(w).Encode(response)
}

// handlePayment handles the /process-payment endpoint
func handlePayment(w http.ResponseWriter, r *http.Request) {
	// Ensure endpoint only accepts POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse and validate the form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	// Extract payment information from form
	paymentToken := r.Form.Get("payment_token")
	billingZip := r.Form.Get("billing_zip")
	amountStr := r.Form.Get("amount")

	// Validate required fields are present
	if paymentToken == "" || billingZip == "" || amountStr == "" {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Details: "Missing required fields",
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Validate and parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil || amount <= 0 {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "VALIDATION_ERROR",
				Details: "Amount must be a positive number",
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	serviceURL := "https://api.pit.paygateway.com"
	endpoint := "/transactions/creditsales"

	// Build request payload
	requestPayload := PaymentAPIRequest{
		ReferenceID: uuid.New().String(),
		Card: CardToken{
			TemporaryToken: paymentToken,
		},
		Customer: Customer{
			BillingAddress: BillingAddress{
				PostalCode: billingZip,
			},
		},
		Payment: Payment{
			Amount:       amountStr,
			CurrencyCode: "840",
		},
		Transaction: TransactionIndicators{
			CountryCode: "840",
			ProcessingIndicators: ProcessingIndicators{
				AllowDuplicate:             true,
				CreateToken:                true,
				AddressVerificationService: true,
			},
		},
	}

	// Marshal request to JSON
	requestBody, err := json.Marshal(requestPayload)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "API_ERROR",
				Details: err.Error(),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", serviceURL+endpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "API_ERROR",
				Details: err.Error(),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Set headers
	// Note: We must bypass Go's automatic header canonicalization for X-GP-* headers
	// because the API requires exact casing (GP not Gp)
	req.Header["Authorization"] = []string{"AuthToken " + createJWT()}
	req.Header["X-GP-Version"] = []string{"2021-04-08"}
	req.Header["X-GP-Api-Key"] = []string{os.Getenv("TRANSACTIONS_API_KEY")}
	req.Header["X-GP-Partner-App-Name"] = []string{"GP Integrated Hosted Fields Sample (Go)"}
	req.Header["Content-Type"] = []string{"application/json"}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "API_ERROR",
				Details: err.Error(),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "API_ERROR",
				Details: err.Error(),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Parse API response
	var apiResponse PaymentAPIResponse
	err = json.Unmarshal(responseBody, &apiResponse)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "API_ERROR",
				Details: err.Error(),
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Check for successful response
	if !strings.EqualFold(apiResponse.Status, "approved") {
		w.Header().Set("Content-Type", "application/json")
		errorResponse := Response{
			Success: false,
			Message: "Payment processing failed",
			Error: &ErrorInfo{
				Code:    "PAYMENT_DECLINED",
				Details: apiResponse.Status,
			},
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	successResponse := Response{
		Success: true,
		Message: fmt.Sprintf("Payment successful! Reference ID: %s", apiResponse.ReferenceID),
		Data: map[string]string{
			"reference_id": apiResponse.ReferenceID,
		},
	}
	json.NewEncoder(w).Encode(successResponse)
}

func main() {
	// Initialize environment configuration
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Set up routes
	http.Handle("/", http.FileServer(http.Dir("static")))
	http.Handle("/config", http.HandlerFunc(handleConfig))
	http.Handle("/process-payment", http.HandlerFunc(handlePayment))

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on http://localhost:%s", port)
	log.Printf("Server also accessible at http://127.0.0.1:%s", port)
	log.Fatal(http.ListenAndServe("0.0.0.0:"+port, nil))
}
