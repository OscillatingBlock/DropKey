# Drop-Key API Documentation

## Overview

Drop-Key is a secure pastebin service that uses Ed25519 cryptographic signatures for authentication and authorization. All pastes are encrypted client-side, and the server only stores ciphertext along with cryptographic signatures for verification.

**Base URL**: `https://yourpasebin.com/api` (configurable via `BASEURL` environment variable)

## Authentication

The API uses JWT (JSON Web Token) authentication for protected endpoints. Authentication is based on Ed25519 digital signatures.

### Authentication Flow
1. Register a user with an Ed25519 public key
2. Authenticate by signing a challenge with your private key
3. Receive a JWT token for accessing protected endpoints
4. Include the token in the `Authorization` header as `Bearer <token>`

## Error Responses

All endpoints return consistent error responses:

```json
{
  "message": "Error description"
}
```

Common HTTP status codes:
- `400` - Bad Request (invalid input, missing required fields)
- `401` - Unauthorized (invalid credentials, missing/invalid token)
- `404` - Not Found (resource doesn't exist)
- `410` - Gone (resource has expired)
- `500` - Internal Server Error

## Endpoints

### User Management

#### Register User
Create a new user account with an Ed25519 public key.

**Endpoint**: `POST /users`

**Authentication**: None required

**Request Body**:
```json
{
  "public_key": "base64-encoded-ed25519-public-key"
}
```

**Response** (201 Created):
```json
{
  "id": "uuid-string"
}
```

**Error Responses**:
- `400` - Empty or invalid public key
- `400` - User already exists (duplicate public key)
- `500` - Internal server error

#### Authenticate User
Authenticate a user by signing a challenge with their private key.

**Endpoint**: `POST /users/auth`

**Authentication**: None required

**Request Body**:
```json
{
  "id": "user-uuid",
  "signature": "base64-encoded-signature",
  "challenge": "base64-encoded-challenge",
  "public_key": "base64-encoded-ed25519-public-key"
}
```

**Response** (200 OK):
```json
{
  "message": "Authentication successful",
  "token": "jwt-token"
}
```

**Error Responses**:
- `400` - Missing user ID, signature, or challenge
- `401` - Invalid signature
- `404` - User not found
- `500` - Internal server error

#### Get User by ID
Retrieve user information by user ID.

**Endpoint**: `GET /users/{id}`

**Authentication**: None required

**Path Parameters**:
- `id` (string, required) - User UUID

**Response** (200 OK):
```json
{
  "id": "user-uuid",
  "public_key": "base64-encoded-public-key"
}
```

**Error Responses**:
- `400` - Missing or invalid user ID
- `404` - User not found
- `500` - Internal server error

#### Get User by Public Key
Retrieve user information by public key.

**Endpoint**: `GET /users?public_key={public_key}`

**Authentication**: None required

**Query Parameters**:
- `public_key` (string, required) - Base64-encoded Ed25519 public key (URL-encoded)

**Response** (200 OK):
```json
{
  "id": "user-uuid",
  "public_key": "base64-encoded-public-key"
}
```

**Error Responses**:
- `400` - Missing or invalid public key
- `404` - User not found
- `500` - Internal server error

### Paste Management

#### Create Paste
Create a new encrypted paste with cryptographic signature.

**Endpoint**: `POST /pastes`

**Authentication**: JWT token required

**Request Body**:
```json
{
  "ciphertext": "base64-encoded-encrypted-content",
  "signature": "base64-encoded-ed25519-signature",
  "public_key": "base64-encoded-ed25519-public-key",
  "expires_in": 3600
}
```

**Fields**:
- `ciphertext` (string, required) - Base64-encoded encrypted content
- `signature` (string, required) - Base64-encoded Ed25519 signature of the ciphertext
- `public_key` (string, required) - Base64-encoded Ed25519 public key (must match authenticated user)
- `expires_in` (integer, required) - Expiration time in seconds from now (max 604800 seconds = 7 days)

**Response** (201 Created):
```json
{
  "id": "paste-uuid",
  "url": "https://yourpasebin.com/paste/uuid#public_key"
}
```

**Error Responses**:
- `400` - Invalid JSON payload, empty ciphertext, invalid signature, etc.
- `401` - Unauthorized access (public key mismatch)
- `500` - Internal server error

#### Get Paste by ID
Retrieve a paste by its ID.

**Endpoint**: `GET /pastes/{id}`

**Authentication**: None required

**Path Parameters**:
- `id` (string, required) - Paste UUID

**Response** (200 OK):
```json
{
  "ID": "paste-uuid",
  "ciphertext": "base64-encoded-encrypted-content",
  "signature": "base64-encoded-signature",
  "public_key": "base64-encoded-public-key",
  "expires_in": "2024-01-01T12:00:00Z"
}
```

**Error Responses**:
- `400` - Invalid paste ID
- `404` - Paste not found
- `410` - Paste has expired
- `500` - Internal server error

#### Update Paste
Update an existing paste (modify content and/or expiration).

**Endpoint**: `PUT /pastes/{id}`

**Authentication**: JWT token required

**Path Parameters**:
- `id` (string, required) - Paste UUID

**Request Body**:
```json
{
  "ciphertext": "base64-encoded-encrypted-content",
  "signature": "base64-encoded-ed25519-signature",
  "public_key": "base64-encoded-ed25519-public-key",
  "expires_in": 3600
}
```

**Response** (200 OK):
```
paste updated
```

**Error Responses**:
- `400` - Invalid input, missing paste ID, invalid expiration time
- `401` - Unauthorized access (public key mismatch)
- `404` - Paste not found
- `500` - Internal server error

#### Get Pastes by Public Key
Retrieve all non-expired pastes for a specific public key.

**Endpoint**: `GET /pastes?public_key={public_key}`

**Authentication**: None required

**Query Parameters**:
- `public_key` (string, required) - Base64-encoded Ed25519 public key (URL-encoded)

**Response** (200 OK):
```json
[
  {
    "ID": "paste-uuid",
    "ciphertext": "base64-encoded-encrypted-content",
    "signature": "base64-encoded-signature",
    "public_key": "base64-encoded-public-key",
    "expires_in": "2024-01-01T12:00:00Z"
  }
]
```

**Error Responses**:
- `400` - Invalid or empty public key
- `401` - Unauthorized access
- `404` - No pastes found or all pastes expired
- `500` - Internal server error

## Security Features

### Cryptographic Signatures
- All pastes must be signed with Ed25519 private keys
- Signatures are verified against the ciphertext content
- Only the owner of the private key can create/modify pastes

### Client-Side Encryption
- All content is encrypted client-side before submission
- The server only stores encrypted ciphertext
- Decryption requires the private key (not stored on server)

### Authentication
- JWT tokens are used for protected endpoints
- Token contains user ID and public key claims
- Public key in requests must match authenticated user

### Expiration
- All pastes have configurable expiration times
- Maximum expiration: 7 days (604800 seconds)
- Expired pastes are automatically removed from responses

## Rate Limiting
- Rate limiting may be implemented (check with service provider)
- Recommended to implement client-side rate limiting for better user experience

## Best Practices

1. **Key Management**: Store private keys securely, never transmit them to the server
2. **Encryption**: Always encrypt sensitive content before submitting
3. **Expiration**: Set appropriate expiration times for your use case
4. **Error Handling**: Implement proper error handling for all API calls
5. **Authentication**: Store JWT tokens securely and refresh as needed

## Examples

### Complete Flow Example (JavaScript)

```javascript
// 1. Generate key pair
const keyPair = crypto.subtle.generateKey(
  { name: "Ed25519", namedCurve: "Ed25519" },
  true,
  ["sign", "verify"]
);

// 2. Register user
const publicKeyBase64 = btoa(publicKeyBytes);
const registerResponse = await fetch('/api/users', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ public_key: publicKeyBase64 })
});

// 3. Authenticate (sign challenge)
const challenge = "your-challenge-here";
const signature = await crypto.subtle.sign("Ed25519", privateKey, challenge);
const authResponse = await fetch('/api/users/auth', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    id: userId,
    signature: btoa(signature),
    challenge: btoa(challenge),
    public_key: publicKeyBase64
  })
});

// 4. Create paste
const ciphertext = encryptContent(plaintext, encryptionKey);
const signature = await crypto.subtle.sign("Ed25519", privateKey, ciphertext);
const pasteResponse = await fetch('/api/pastes', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': `Bearer ${jwtToken}`
  },
  body: JSON.stringify({
    ciphertext: btoa(ciphertext),
    signature: btoa(signature),
    public_key: publicKeyBase64,
    expires_in: 3600
  })
});
```

## Environment Variables

- `BASEURL` - Base URL for the service (default: "https://yourpasebin.com")
- `JWTSECRET` - Secret key for JWT token signing (required)

## Version

This documentation is for the current version of the Drop-Key API. Check the service for any updates or changes to the API specification.
