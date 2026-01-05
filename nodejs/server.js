/**
 * Card Payment Processing Server - Node.js
 *
 * This server demonstrates card payment processing using Global Payments hosted fields
 * tokenization and JWT authentication. It handles tokenized card data and billing
 * information to process payments securely through the Global Payments API via direct
 * HTTP requests (not SDK-based).
 */

import express from 'express';
import * as dotenv from 'dotenv';
import jwt from 'jsonwebtoken';
import { randomUUID } from 'crypto';

// Load environment variables from .env file
dotenv.config();

/**
 * Initialize Express application with necessary middleware
 */
const app = express();
const port = process.env.PORT || 8000;

app.use(express.static('.')); // Serve static files
app.use(express.urlencoded({ extended: true })); // Parse form data
app.use(express.json()); // Parse JSON requests

/**
 * Create JWT for authentication
 *
 * @returns {string} JWT token
 */
function createJWT() {
    const key = process.env.AUTHTOKEN_JWT_SECRET;

    const payload = {
        type: 'AuthTokenV2',
        region: 'US',
        account_credential: process.env.ACCOUNT_CREDENTIAL,
        ts: Date.now()
    };

    return jwt.sign(payload, key, { algorithm: 'HS256', noTimestamp: true });
}

/**
 * Sanitize postal code by removing invalid characters
 *
 * @param {string|null} postalCode The postal code to sanitize
 * @returns {string} Sanitized postal code containing only alphanumeric
 *                   characters and hyphens, limited to 10 characters
 */
function sanitizePostalCode(postalCode) {
    if (!postalCode) {
        return '';
    }

    const sanitized = postalCode.replace(/[^a-zA-Z0-9-]/g, '');
    return sanitized.slice(0, 10);
}

/**
 * Config endpoint - provides hosted fields API key for client-side initialization
 */
app.get('/config', (req, res) => {
    res.json({
        success: true,
        data: {
            apiKey: process.env.HOSTED_FIELDS_API_KEY
        }
    });
});

/**
 * Payment processing endpoint
 */
app.post('/process-payment', async (req, res) => {
    try {
        // Validate required fields
        if (!req.body.payment_token || !req.body.billing_zip || !req.body.amount) {
            throw new Error('Missing required fields');
        }

        // Parse and validate amount
        const amount = parseFloat(req.body.amount);
        if (amount <= 0) {
            throw new Error('Invalid amount');
        }

        const serviceUrl = 'https://api.pit.paygateway.com';
        const endpoint = '/transactions/creditsales';

        // Initialize payment data using tokenized card information
        const requestHeaders = {
            'Authorization': 'AuthToken ' + createJWT(),
            'X-GP-Version': '2021-04-08',
            'X-GP-Api-Key': process.env.TRANSACTIONS_API_KEY,
            'X-GP-Partner-App-Name': 'GP Integrated Hosted Fields Sample (Node.js)',
            'Content-Type': 'application/json'
        };

        const requestBody = {
            reference_id: randomUUID(),
            card: { temporary_token: req.body.payment_token },
            customer: {
                billing_address: {
                    postal_code: req.body.billing_zip
                }
            },
            payment: {
                amount: req.body.amount,
                currency_code: '840'
            },
            transaction: {
                country_code: '840',
                processing_indicators: {
                    allow_duplicate: true,
                    create_token: true,
                    address_verification_service: true
                }
            }
        };

        const response = await fetch(serviceUrl + endpoint, {
            method: 'POST',
            headers: requestHeaders,
            body: JSON.stringify(requestBody)
        });

        const data = await response.json();

        // Verify transaction was successful
        if (data.status !== 'approved') {
            res.status(400).json({
                success: false,
                message: 'Payment processing failed',
                error: {
                    code: 'PAYMENT_DECLINED',
                    details: data.status
                }
            });
            return;
        }

        // Return success response with transaction ID
        res.json({
            success: true,
            message: 'Payment successful! Reference ID: ' + data.reference_id,
            data: {
                reference_id: data.reference_id
            }
        });
    } catch (error) {
        // Handle payment processing errors
        res.status(400).json({
            success: false,
            message: 'Payment processing failed',
            error: {
                code: 'API_ERROR',
                details: error.message
            }
        });
    }
});

// Start the server
app.listen(port, '0.0.0.0', () => {
    console.log(`Server running at http://localhost:${port}`);
});
