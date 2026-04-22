# ezauth

A lightweight, language-agnostic authentication server for local development. Drop it alongside any project and get a working auth API in seconds — no libraries, no boilerplate, no lock-in.

## How it works

ezauth runs as a standalone HTTP server. Your application talks to it over REST, so it doesn't matter what language or framework you're using. Register users, log them in, protect routes — all via simple JSON API calls.

## Installation

Download the latest binary for your platform from the [Releases page](https://github.com/ConnorBrightman/ezauth/releases) and place it in your project folder.

| Platform | File |
|----------|------|
| Windows | `ezauth.exe` |
| macOS (Apple Silicon) | `ezauth-mac-arm` |
| macOS (Intel) | `ezauth-mac` |
| Linux | `ezauth-linux` |

On macOS/Linux, make it executable after downloading:
```bash
chmod +x ezauth-mac-arm  # adjust to match the filename you downloaded
```

> **Note:** The downloaded file keeps its original name (e.g. `ezauth-mac-arm`). Commands below use `ezauth` for brevity — substitute the actual filename if you haven't renamed it or added it to your PATH. For example:
> ```bash
> ./ezauth-mac-arm start
> ```

**Optionally rename it** to `ezauth` (or `ezauth.exe` on Windows) so the commands below work as written:
```bash
mv ezauth-mac-arm ezauth   # macOS/Linux
```

**Or build from source** (requires Go 1.21+):

```bash
git clone https://github.com/ConnorBrightman/ezauth.git
cd ezauth
go install ./cmd/ezauth
```

---

## Quick start

Place the binary in your project folder, then open a terminal there.

**Zero config** — just run it:

```bash
./ezauth start          # macOS/Linux
ezauth.exe start        # Windows
```

The server starts on `http://127.0.0.1:8080`. Users are stored in memory and sessions are lost on restart. Good for quick testing.

**Persistent setup** — run this once in your project directory:

```bash
./ezauth init   # generates config.yaml with a unique JWT secret and SQLite storage
./ezauth start
```

`init` will prompt you to choose a storage backend:

| Option | Backend | Best for |
|--------|---------|----------|
| `1` | Memory | Quick testing — data lost on restart |
| `2` | File (JSON) | Simple persistence, human-readable |
| `3` | SQLite | Local development with persistent storage (default) |
| `4` | PostgreSQL | External Postgres server |
| `5` | MySQL | External MySQL server |

For Postgres and MySQL you'll be prompted for a connection string (DSN). All other backends store data locally in your project folder.

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
