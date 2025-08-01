
# DropKey

A minimal, end-to-end encrypted pastebin built with Go. This service enables users to create, share, and fetch encrypted messages—**no accounts, emails, or passwords required**.

## 🔐 Key Features

- **End-to-End Encryption (E2EE)**  
  All messages are encrypted on the client side with the user’s public key. The server only stores ciphertext—never plain text.

- **Public Key Authentication**  
  Users authenticate using challenge–response signatures via Ed25519 public/private key pairs.

- **Stateless & Private**  
  The backend stores no personal user data. Only public keys and signed challenges are processed.

- **JWT-Based Sessions**  
  On successful auth, users receive a JWT to access protected routes like creating or updating pastes.

- **Signed Pastes**  
  Every paste is signed with the user’s private key to ensure authenticity and integrity.

- **Expiring Pastes**  
  Users can optionally set pastes to expire. Expired entries are automatically excluded from responses.

---

## 📁 Project Structure

```

DropKey/
├── cmd
│   └── api
│       └── main.go            # Entry point
├── internal
│   ├── paste
│   │   ├── paste.go           # Paste model
│   │   ├── handler.go         # Paste HTTP handlers
│   │   ├── service.go         # Paste business logic
│   │   └── repository.go      # Paste database operations
│   ├── user
│   │   ├── user.go            # User model
│   │   ├── handler.go         # User HTTP handlers
│   │   ├── service.go         # User business logic
│   │   └── repository.go      # User database operations
│   ├── db
│   │   └── db.go              # Bun + MySQL setup
│   ├── router
│   │   └── router.go          # Route definitions
│   ├── middleware
│   │   └── middleware.go      # Logging, auth middleware
│   └── utils
│       └── errors.go          # Custom error types
├── .env                       # Environment variables
├── go.mod                     # Go dependencies
└── README.md

````

---

## Tech Stack

- **Language:** Go 1.21+
- **Framework:** [Echo](https://echo.labstack.com)
- **ORM:** [Bun](https://bun.uptrace.dev) with MySQL
- **Authentication:** Ed25519 signatures + JWT
- **Logging:** `log/slog`

---

## Getting Started

```bash
git clone https://github.com/OscillatingBlock/DropKey.git
cd DropKey
sudo docker-compose up --build
````

The API server will be available at:
`http://localhost:8081`

You can now interact with it using tools like `curl` or Postman. Refer to [`API_DOCS.md`](./API_DOCS.md) for available endpoints.

---

## 🌐 Environment Variables

Create a `.env` file in `cmd/api/` with the following variables:

```env
PORT=8081
DB_HOST=dropkey-mysql
DB_PORT=3306
DB_USER=root
DB_PASSWORD=your_password
DB_NAME=dropkey
JWT_SECRET=your_super_secret_jwt_key
```

> Note: These variables are automatically loaded using `github.com/joho/godotenv`.

---

## Notes

* The **server never sees decrypted content**.
* Paste decryption is **client-only** using the key stored in the URL fragment (`#hash`)—which is **not** sent to the server.

---

## Related Docs

* [📄 API Reference →](./API_DOCS.md)

---

## Contributing

PRs are welcome. Please keep the codebase clean, modular, and idiomatic Go. Use `go fmt` and run tests if added.

---
