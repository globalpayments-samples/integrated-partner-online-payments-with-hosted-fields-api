"""
Card Payment Processing Server - Python Flask

This server demonstrates card payment processing using Global Payments hosted fields
tokenization and JWT authentication. It handles tokenized card data and billing
information to process payments securely through the Global Payments API via direct
HTTP requests (not SDK-based).
"""

import os
import re
import json
import time
import uuid
import requests
from flask import Flask, request, jsonify
from dotenv import load_dotenv
import jwt

# Load environment variables
load_dotenv()

# Initialize application
app = Flask(__name__, static_folder='.')


def create_jwt():
    """
    Create JWT for authentication

    Returns:
        str: JWT token
    """
    key = os.getenv('AUTHTOKEN_JWT_SECRET')

    payload = {
        'type': 'AuthTokenV2',
        'region': 'US',
        'account_credential': os.getenv('ACCOUNT_CREDENTIAL'),
        'ts': int(time.time() * 1000)
    }

    return jwt.encode(payload, key, algorithm='HS256')


def sanitize_postal_code(postal_code):
    """
    Sanitize postal code by removing invalid characters

    Args:
        postal_code: The postal code to sanitize

    Returns:
        str: Sanitized postal code containing only alphanumeric
             characters and hyphens, limited to 10 characters
    """
    if not postal_code:
        return ''

    sanitized = re.sub(r'[^a-zA-Z0-9-]', '', postal_code)
    return sanitized[:10]


@app.route('/')
def index():
    """Serve the main HTML page."""
    return app.send_static_file('index.html')


@app.route('/config')
def get_config():
    """Config endpoint - provides hosted fields API key for client-side initialization."""
    return jsonify({
        'success': True,
        'data': {
            'apiKey': os.getenv('HOSTED_FIELDS_API_KEY')
        }
    })


@app.route('/process-payment', methods=['POST'])
def process_payment():
    """Payment processing endpoint."""
    try:
        # Validate required fields
        if not request.form.get('payment_token') or not request.form.get('billing_zip') or not request.form.get('amount'):
            raise Exception('Missing required fields')

        # Parse and validate amount
        amount = float(request.form.get('amount'))
        if amount <= 0:
            raise Exception('Invalid amount')

        service_url = 'https://api.pit.paygateway.com'
        endpoint = '/transactions/creditsales'

        # Initialize payment data using tokenized card information
        request_headers = {
            'Authorization': 'AuthToken ' + create_jwt(),
            'X-GP-Version': '2021-04-08',
            'X-GP-Api-Key': os.getenv('TRANSACTIONS_API_KEY'),
            'X-GP-Partner-App-Name': 'GP Integrated Hosted Fields Sample (Python)',
            'Content-Type': 'application/json'
        }

        request_body = {
            'reference_id': str(uuid.uuid4()),
            'card': {'temporary_token': request.form.get('payment_token')},
            'customer': {
                'billing_address': {
                    'postal_code': request.form.get('billing_zip')
                }
            },
            'payment': {
                'amount': request.form.get('amount'),
                'currency_code': '840'
            },
            'transaction': {
                'country_code': '840',
                'processing_indicators': {
                    'allow_duplicate': True,
                    'create_token': True,
                    'address_verification_service': True
                }
            }
        }

        response = requests.post(
            service_url + endpoint,
            headers=request_headers,
            json=request_body,
            timeout=65
        )

        data = response.json()

        # Verify transaction was successful
        if data.get('status') != 'approved':
            return jsonify({
                'success': False,
                'message': 'Payment processing failed',
                'error': {
                    'code': 'PAYMENT_DECLINED',
                    'details': data.get('status')
                }
            }), 400

        # Return success response with transaction ID
        return jsonify({
            'success': True,
            'message': f"Payment successful! Reference ID: {data.get('reference_id')}",
            'data': {
                'reference_id': data.get('reference_id')
            }
        })

    except Exception as e:
        # Handle payment processing errors
        return jsonify({
            'success': False,
            'message': 'Payment processing failed',
            'error': {
                'code': 'API_ERROR',
                'details': str(e)
            }
        }), 400


# Start the server if this file is run directly
if __name__ == '__main__':
    port = int(os.getenv('PORT', 8000))
    print(f"Server running at http://localhost:{port}")
    app.run(host='0.0.0.0', port=port, debug=True)
