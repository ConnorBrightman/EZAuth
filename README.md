# ezauth

A lightweight, language-agnostic authentication server for local development. Drop it alongside any project and get a working auth API in seconds — no libraries, no boilerplate, no lock-in.

## How it works

ezauth runs as a standalone HTTP server. Your application talks to it over REST, so it doesn't matter what language or framework you're using. Register users, log them in, protect routes — all via simple JSON API calls.

## Installation

**Build from source** (requires Go 1.21+):

```bash
git clone https://github.com/ConnorBrightman/ezauth.git
cd ezauth
go install ./cmd/ezauth
```

## Quick start

```bash
# 1. Initialise — creates config.yaml and data directory in the current folder
ezauth init

# 2. Start the server
ezauth start
```

The server starts on `http://127.0.0.1:8080` by default.

## Configuration

`ezauth init` generates a `config.yaml` in the current directory. Edit it to suit your project:

```yaml
host: 127.0.0.1
port: "8080"
jwt_secret: super-secret-key        # Change this in any real environment
access_token_expiry: 5m             # Go duration string
refresh_token_expiry: 168h          # 7 days
storage: file                       # "file" or "memory"
file_path: ezauth-data/users.json   # Only used when storage: file
logging_enabled: true
```

All values can also be overridden with environment variables (e.g. `JWT_SECRET=mysecret ezauth start`).

| Key | Default | Description |
|-----|---------|-------------|
| `host` | `127.0.0.1` | Address to listen on |
| `port` | `8080` | Port to listen on |
| `jwt_secret` | `super-secret-key` | HMAC secret for signing JWTs |
| `access_token_expiry` | `5m` | Access token lifetime |
| `refresh_token_expiry` | `168h` | Refresh token lifetime |
| `storage` | `memory` | `sqlite`, `file`, or `memory` |
| `database_path` | `ezauth-data/ezauth.db` | Path for SQLite database file |
| `file_path` | `ezauth-data/users.json` | Path for file storage |
| `logging_enabled` | `true` | Log incoming requests |

## API reference

All endpoints return JSON in the following envelope:

```json
{ "success": true, "data": { ... } }
{ "success": false, "error": { "code": "ERROR_CODE", "message": "human readable" } }
```

---

### `GET /health`

Check that the server is running.

**Response `200`**
```json
{
  "success": true,
  "data": { "status": "healthy", "version": "1.0.0" }
}
```

---

### `POST /auth/register`

Create a new user account.

**Request**
```json
{ "email": "user@example.com", "password": "mypassword" }
```

**Response `201`**
```json
{
  "success": true,
  "data": { "message": "user registered successfully" }
}
```

**Errors**

| Status | Code | Reason |
|--------|------|--------|
| `400` | `BAD_REQUEST` | Missing email or password |
| `409` | `EMAIL_ALREADY_EXISTS` | That email is already registered |

---

### `POST /auth/login`

Authenticate a user and receive tokens.

**Request**
```json
{ "email": "user@example.com", "password": "mypassword" }
```

**Response `200`**
```json
{
  "success": true,
  "data": {
    "access_token": "<jwt>",
    "refresh_token": "<opaque token>"
  }
}
```

**Errors**

| Status | Code | Reason |
|--------|------|--------|
| `401` | `INVALID_CREDENTIALS` | Email not found or wrong password |

---

### `GET /auth/me`

Return the currently authenticated user. Requires a valid access token.

**Headers**
```
Authorization: Bearer <access_token>
```

**Response `200`**
```json
{
  "success": true,
  "data": { "user_id": "abc-123", "email": "user@example.com" }
}
```

**Errors**

| Status | Reason |
|--------|--------|
| `401` | Missing, invalid, or expired token |

---

### `POST /auth/logout`

Invalidates the current session by clearing the user's refresh token. Requires a valid access token.

**Headers**
```
Authorization: Bearer <access_token>
```

**Response `200`**
```json
{
  "success": true,
  "data": { "message": "logged out successfully" }
}
```

**Errors**

| Status | Reason |
|--------|--------|
| `401` | Missing or invalid access token |

---

### `POST /auth/refresh`

Exchange a refresh token for a new access token and a new refresh token. The old refresh token is invalidated immediately (token rotation).

**Request**
```json
{ "email": "user@example.com", "refresh_token": "<refresh_token>" }
```

**Response `200`**
```json
{
  "success": true,
  "data": {
    "access_token": "<new jwt>",
    "refresh_token": "<new opaque token>"
  }
}
```

**Errors**

| Status | Code | Reason |
|--------|------|--------|
| `401` | `UNAUTHORIZED` | Invalid, expired, or already-used refresh token |

---

## Client examples

Because ezauth is a plain HTTP API, it works with any language or tool.

**curl**
```bash
# Register
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"mypassword"}'

# Login
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"mypassword"}'

# Protected route (replace <token> with the access_token from login)
curl -s http://localhost:8080/auth/me \
  -H "Authorization: Bearer <token>"
```

**JavaScript (fetch)**
```js
const base = 'http://localhost:8080';

const { data } = await fetch(`${base}/auth/login`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email: 'user@example.com', password: 'mypassword' }),
}).then(r => r.json());

const { access_token, refresh_token } = data;
```

**Python (requests)**
```python
import requests

base = 'http://localhost:8080'

res = requests.post(f'{base}/auth/login', json={
    'email': 'user@example.com',
    'password': 'mypassword',
})
tokens = res.json()['data']
access_token = tokens['access_token']
```

## Running the tests

```bash
go test ./...
```
