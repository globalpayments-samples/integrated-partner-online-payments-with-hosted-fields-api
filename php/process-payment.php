<?php

declare(strict_types=1);

/**
 * Card Payment Processing Script
 *
 * This script demonstrates card payment processing using the Global Payments SDK.
 * It handles tokenized card data and billing information to process payments
 * securely through the Global Payments API.
 *
 * PHP version 7.4 or higher
 *
 * @category  Payment_Processing
 * @package   GlobalPayments_Sample
 * @author    Global Payments
 * @license   MIT License
 * @link      https://github.com/globalpayments
 */

require_once 'vendor/autoload.php';
$curl = require_once 'curl.php';

use Dotenv\Dotenv;
use Firebase\JWT\JWT;

ini_set('display_errors', '0');

/**
 * Configure the SDK
 *
 * Sets up the Global Payments SDK with necessary credentials and settings
 * loaded from environment variables.
 *
 * @return void
 */
function configureSdk(): void
{
    $dotenv = Dotenv::createImmutable(__DIR__);
    $dotenv->load();
}

function createJWT(): string
{
    $key = $_ENV['AUTHTOKEN_JWT_SECRET'];

    $payload = [
        'type' => 'AuthTokenV2',
        'region' => 'US',
        'account_credential' => $_ENV['ACCOUNT_CREDENTIAL'],
        'ts' => floor(microtime(true) * 1000),
    ];

    return JWT::encode($payload, $key, 'HS256');
}

function getGuid()
{
    if (function_exists('com_create_guid')) {
        return trim(com_create_guid(), '{}');
    }

    $data = openssl_random_pseudo_bytes(16);
    $data[6] = chr(ord($data[6]) & 0x0f | 0x40); // set version to 0100
    $data[8] = chr(ord($data[8]) & 0x3f | 0x80); // set bits 6-7 to 10
    return vsprintf('%s%s-%s-%s-%s-%s%s%s', str_split(bin2hex($data), 4));
}

/**
 * Sanitize postal code by removing invalid characters
 *
 * @param string|null $postalCode The postal code to sanitize
 *
 * @return string Sanitized postal code containing only alphanumeric
 *                characters and hyphens, limited to 10 characters
 */
function sanitizePostalCode(?string $postalCode): string
{
    if ($postalCode === null) {
        return '';
    }
    
    $sanitized = preg_replace('/[^a-zA-Z0-9-]/', '', $postalCode);
    return substr($sanitized, 0, 10);
}

// Initialize SDK configuration
configureSdk();

try {
    // Validate required fields
    if (!isset($_POST['payment_token'], $_POST['billing_zip'], $_POST['amount'])) {
        throw new Exception('Missing required fields');
    }
    
    // Parse and validate amount
    $amount = floatval($_POST['amount']);
    if ($amount <= 0) {
        throw new Exception('Invalid amount');
    }

    $serviceUrl = 'https://api.pit.paygateway.com';
    $endpoint = '/transactions/creditsales';
    $verb = 'POST';
    $queryString = '';
    $contentType = 'application/json';
    $timeout = 65000;

    // Initialize payment data using tokenized card information
    $requestHeaders = [
        'Authorization' => 'AuthToken ' . createJWT(),
        // 'X-GP-Request-Id' => $requestId,
        'X-GP-Version' => '2021-04-08',
        'X-GP-Api-Key' => $_ENV['TRANSACTIONS_API_KEY'],
        'X-GP-Partner-App-Name' => 'GP Integrated Hosted Fields Sample (PHP)',
    ];

    $request = [
        'reference_id' => getGuid(),
        'card' => [ 'temporary_token' => $_POST['payment_token'] ],
        'customer' => [
            'billing_address' => [
                'postal_code' => $_POST['billing_zip'],
            ],
        ],
        'payment' => [
            'amount' => $_POST['amount'],
            'currency_code' => '840',
        ],
        'transaction' => [
            'country_code' => '840',
            'processing_indicators' => [
                'allow_duplicate' => true,
                'create_token' => true,
                'address_verification_service' => true,
            ],
        ],
    ];

    $rawResponse = $curl(
        $serviceUrl, 
        $endpoint, 
        $queryString, 
        $requestHeaders, 
        json_encode($request), 
        $verb, 
        $contentType, 
        $timeout
    );

    $response = json_decode($rawResponse[0]);
    
    // Verify transaction was successful
    if ($response->status !== 'approved') {
        http_response_code(400);
        echo json_encode([
            'success' => false,
            'message' => 'Payment processing failed',
            'error' => [
                'code' => 'PAYMENT_DECLINED',
                'details' => $response->status
            ]
        ]);
        exit;
    }

    // Return success response with transaction ID
    echo json_encode([
        'success' => true,
        'message' => 'Payment successful! Reference ID: ' . $response->reference_id,
        'data' => [
            'reference_id' => $response->reference_id
        ]
    ]);
} catch (Exception $e) {
    // Handle payment processing errors
    http_response_code(400);
    echo json_encode([
        'success' => false,
        'message' => 'Payment processing failed',
        'error' => [
            'code' => 'API_ERROR',
            'details' => $e->getMessage()
        ]
    ]);
}
