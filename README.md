# pkce-poc

A minimal Go CLI and local API server demonstrating the OAuth 2.0 Authorization Code flow with PKCE, using Auth0 for authentication.

## What it does

- `login` — Opens the browser to Auth0, handles the callback, exchanges the authorization code for tokens using PKCE, and stores the access token in the OS keychain (macOS Keychain, Linux secret service, Windows Credential Manager).
- `fetch` — Reads the stored token and calls the protected local API endpoint.
- `serve` — Starts a local HTTP server on port 8080 that validates Auth0-issued JWTs and returns protected data.

## Project structure

```
pkce-poc/
├── cmd/
│   ├── root.go        # cobra root command
│   ├── login.go       # "login" subcommand
│   ├── logout.go      # "logout" subcommand
│   ├── fetch.go       # "fetch" subcommand
│   └── serve.go       # "serve" subcommand
├── internal/
│   ├── auth/
│   │   ├── pkce.go    # PKCE code verifier + S256 challenge generation
│   │   ├── flow.go    # login/logout flow: build auth URL, callback server, token exchange
│   │   └── token.go   # keychain-based token storage + JWT expiry check + refresh
│   └── api/
│       └── server.go  # JWT-validating HTTP server, serves protected endpoint
├── main.go
└── go.mod
```

## Configuration

See [docs/auth0-setup.md](docs/auth0-setup.md) for a full Auth0 setup walkthrough.

Once Auth0 is configured, set your tenant domain and client ID in two files:

**`internal/auth/flow.go`**
```go
Auth0Domain = "YOUR_AUTH0_TENANT.us.auth0.com"
ClientID    = "YOUR_CLIENT_ID"
```

**`internal/api/server.go`**
```go
auth0Domain = "YOUR_AUTH0_TENANT.us.auth0.com"
```

The Auth0 application must be configured as a **Native** type (no client secret) with:
- Allowed Callback URL: `http://localhost:8085/callback`
- Allowed Logout URL: `http://localhost:8085`

The Auth0 API must be configured with:
- Audience/Identifier: `http://localhost:8080/api`
- Signing Algorithm: RS256

## Usage

```bash
# Build
go build -o pkce-poc .

# Terminal 1: start the API server
./pkce-poc serve

# Terminal 2: authenticate (opens browser)
./pkce-poc login

# Terminal 2: call the protected endpoint
./pkce-poc fetch
# Output: {"friend":"Marco"}

# Log out (clears keychain + opens Auth0 logout)
./pkce-poc logout
```

## Design note

The CLI and API server are intentionally combined into a single binary for simplicity. In a real-world setup these would be separate services — the CLI distributed to users, the API deployed independently.

## Key implementation notes

- The PKCE **code verifier** is sent during token exchange — not the challenge. The challenge is only used in the initial authorize URL.
- The `state` parameter is validated on callback to prevent CSRF.
- The `audience` parameter in the authorize URL is required; without it Auth0 returns an opaque token that the API cannot validate.
- The Auth0 issuer URL for JWT validation must include a trailing slash: `https://YOUR_TENANT.us.auth0.com/`.
- Tokens are stored in the OS keychain via `github.com/zalando/go-keyring`.
- The `fetch` command checks token expiry and attempts a refresh before calling the API.
