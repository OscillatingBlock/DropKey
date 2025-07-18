# KeyDrop 

A minimal, end-to-end encrypted pastebin built with Go. This service allows users to create, share, and fetch encrypted messages without requiring email, passwords, or accounts.

## Key Features

- **End-to-End Encryption (E2EE)**  
  All messages are encrypted client-side using the user's public key. The server only stores ciphertext, never plain text.

- **Public Key-Based Authentication**  
  Authentication uses challenge–response signatures with public/private key pairs. No passwords, emails, or usernames are involved.

- **Stateless and Private**  
  No personal user data is stored. Only public keys and signed challenges are used to validate requests.

- **JWT-Based Session Tokens**  
  After authentication, users receive a JWT that can be used to access protected endpoints (e.g., creating or updating pastes).

- **Signed Pastes**  
  Each paste is signed using the user's private key. This guarantees authenticity and integrity of content.

- **Expiring Content**  
  Pastes can be set to expire after a certain time. Expired pastes are not returned in responses.

## Project Structure

```

KeyDrop/
├── cmd
│   └── api
│       └── main.go              # Entry point
├── internal
│   ├── paste                   # Paste domain
│   │   ├── paste.go            # Paste model
│   │   ├── handler.go          # Paste HTTP handlers
│   │   ├── service.go          # Paste business logic
│   │   └── repository.go       # Paste data access
│   ├── user                    # User domain
│   │   ├── user.go             # User model
│   │   ├── handler.go          # User HTTP handlers
│   │   ├── service.go          # User business logic
│   │   └── repository.go       # User data access
│   ├── db
│   │   └── db.go               # Database setup (Bun ORM)
│   ├── router
│   │   └── router.go           # Route definitions
│   ├── middleware
│   │   └── middleware.go       # Logging, authentication middleware
│   └── utils
│       ├── errors.go           # Custom errors
├── go.mod                      # Go module dependencies
└── README.md                   # Project documentation

````

## Tech Stack

- **Language:** Go 1.21+
- **Framework:** Echo
- **ORM:** Bun (with MySql)
- **Auth:** JWT + Public Key Cryptography
- **Logger:** `log/slog`

## Note

- The server **never** sees or stores plain text.
- Decryption happens **only** client-side using the public key hash in the fragment (`#`) portion of the paste URL.

## Related

- [Endpoints documentation →](./API_DOCS.md) *REFER THIS FOR API DOCS*

---

```bash
git clone https://github.com/yourname/pastebin-service.git
cd pastebin-service
go run ./cmd/api
````
