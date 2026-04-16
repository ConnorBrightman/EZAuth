package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	api "github.com/ConnorBrightman/ezauth/internal/api"
	"github.com/ConnorBrightman/ezauth/internal/auth"
)

const (
	testEmail    = "test@example.com"
	testPassword = "password123"
)

var testSecret = []byte("test-secret")

// newTestServer creates a fresh server backed by an in-memory repository.
func newTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	return newTestServerWithRepo(t, auth.NewMemoryUserRepository())
}

// newSQLiteTestServer creates a fresh server backed by a temporary SQLite database.
func newSQLiteTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "test.db")
	repo, err := auth.NewSQLiteUserRepository(path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { repo.Close() })
	return newTestServerWithRepo(t, repo)
}

func newTestServerWithRepo(t *testing.T, repo auth.UserRepository) *httptest.Server {
	t.Helper()
	service := auth.NewService(repo, testSecret, 5*time.Minute, 168*time.Hour)
	srv := httptest.NewServer(api.NewRouter(service, testSecret))
	t.Cleanup(srv.Close)
	return srv
}

// doRequest sends an HTTP request to the test server.
func doRequest(t *testing.T, srv *httptest.Server, method, path string, body any, token string) *http.Response {
	t.Helper()

	var buf *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		buf = bytes.NewBuffer(b)
	} else {
		buf = &bytes.Buffer{}
	}

	req, err := http.NewRequest(method, srv.URL+path, buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

type apiResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   *apiError       `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func decodeResponse(t *testing.T, resp *http.Response) apiResponse {
	t.Helper()
	defer resp.Body.Close()
	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return r
}

// registerAndLogin registers testEmail and returns both tokens.
func registerAndLogin(t *testing.T, srv *httptest.Server) (accessToken, refreshToken string) {
	t.Helper()
	creds := map[string]string{"email": testEmail, "password": testPassword}

	reg := doRequest(t, srv, http.MethodPost, "/auth/register", creds, "")
	if reg.StatusCode != http.StatusCreated {
		reg.Body.Close()
		t.Fatalf("setup register: expected 201, got %d", reg.StatusCode)
	}
	reg.Body.Close()

	login := doRequest(t, srv, http.MethodPost, "/auth/login", creds, "")
	if login.StatusCode != http.StatusOK {
		login.Body.Close()
		t.Fatalf("setup login: expected 200, got %d", login.StatusCode)
	}
	r := decodeResponse(t, login)

	var tokens map[string]string
	if err := json.Unmarshal(r.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	return tokens["access_token"], tokens["refresh_token"]
}

// --- SQLite backend smoke test ---

// TestSQLite_FullAuthFlow runs the happy path against a real SQLite database
// to verify the backend is wired up correctly.
func TestSQLite_FullAuthFlow(t *testing.T) {
	srv := newSQLiteTestServer(t)
	creds := map[string]string{"email": testEmail, "password": testPassword}

	// Register
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", creds, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	// Login
	resp = doRequest(t, srv, http.MethodPost, "/auth/login", creds, "")
	r := decodeResponse(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", resp.StatusCode)
	}
	var tokens map[string]string
	if err := json.Unmarshal(r.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}

	// Me
	resp = doRequest(t, srv, http.MethodGet, "/auth/me", nil, tokens["access_token"])
	r = decodeResponse(t, resp)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("me: expected 200, got %d", resp.StatusCode)
	}
	var user map[string]string
	if err := json.Unmarshal(r.Data, &user); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	if user["email"] != testEmail {
		t.Errorf("expected email %q, got %q", testEmail, user["email"])
	}

	// Refresh
	body := map[string]string{"email": testEmail, "refresh_token": tokens["refresh_token"]}
	resp = doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("refresh: expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// --- Health ---

func TestHealth_OK(t *testing.T) {
	srv := newTestServer(t)
	resp := doRequest(t, srv, http.MethodGet, "/health", nil, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}
}

func TestHealth_WrongMethod(t *testing.T) {
	srv := newTestServer(t)
	resp := doRequest(t, srv, http.MethodPost, "/health", nil, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "METHOD_NOT_ALLOWED" {
		t.Errorf("expected METHOD_NOT_ALLOWED error code, got %v", r.Error)
	}
}

// --- Register ---

func TestRegister_Success(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": testEmail, "password": testPassword}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}
}

func TestRegister_DuplicateEmail(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": testEmail, "password": testPassword}

	first := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	first.Body.Close()

	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("expected 409, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "EMAIL_ALREADY_EXISTS" {
		t.Errorf("expected EMAIL_ALREADY_EXISTS, got %v", r.Error)
	}
}

func TestRegister_MissingEmail(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": "", "password": testPassword}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_EMAIL" {
		t.Errorf("expected INVALID_EMAIL, got %v", r.Error)
	}
}

func TestRegister_MissingPassword(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": testEmail, "password": ""}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_PASSWORD" {
		t.Errorf("expected INVALID_PASSWORD, got %v", r.Error)
	}
}

func TestRegister_InvalidEmail(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": "notanemail", "password": testPassword}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_EMAIL" {
		t.Errorf("expected INVALID_EMAIL, got %v", r.Error)
	}
}

func TestRegister_PasswordTooShort(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": testEmail, "password": "short"}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_PASSWORD" {
		t.Errorf("expected INVALID_PASSWORD, got %v", r.Error)
	}
}

func TestRegister_PasswordTooLong(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": testEmail, "password": strings.Repeat("a", 73)}
	resp := doRequest(t, srv, http.MethodPost, "/auth/register", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_PASSWORD" {
		t.Errorf("expected INVALID_PASSWORD, got %v", r.Error)
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	srv := newTestServer(t)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/auth/register", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "BAD_REQUEST" {
		t.Errorf("expected BAD_REQUEST, got %v", r.Error)
	}
}

// --- Login ---

func TestLogin_Success(t *testing.T) {
	srv := newTestServer(t)
	creds := map[string]string{"email": testEmail, "password": testPassword}

	reg := doRequest(t, srv, http.MethodPost, "/auth/register", creds, "")
	reg.Body.Close()

	resp := doRequest(t, srv, http.MethodPost, "/auth/login", creds, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}

	var tokens map[string]string
	if err := json.Unmarshal(r.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	if tokens["access_token"] == "" {
		t.Error("expected non-empty access_token")
	}
	if tokens["refresh_token"] == "" {
		t.Error("expected non-empty refresh_token")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	srv := newTestServer(t)
	creds := map[string]string{"email": testEmail, "password": testPassword}

	reg := doRequest(t, srv, http.MethodPost, "/auth/register", creds, "")
	reg.Body.Close()

	wrong := map[string]string{"email": testEmail, "password": "wrongpassword"}
	resp := doRequest(t, srv, http.MethodPost, "/auth/login", wrong, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got %v", r.Error)
	}
}

func TestLogin_UnknownEmail(t *testing.T) {
	srv := newTestServer(t)
	body := map[string]string{"email": "nobody@example.com", "password": testPassword}
	resp := doRequest(t, srv, http.MethodPost, "/auth/login", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "INVALID_CREDENTIALS" {
		t.Errorf("expected INVALID_CREDENTIALS, got %v", r.Error)
	}
}

// --- Me ---

func TestMe_Authorized(t *testing.T) {
	srv := newTestServer(t)
	accessToken, _ := registerAndLogin(t, srv)

	resp := doRequest(t, srv, http.MethodGet, "/auth/me", nil, accessToken)
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}

	var user map[string]string
	if err := json.Unmarshal(r.Data, &user); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}
	if user["email"] != testEmail {
		t.Errorf("expected email %q, got %q", testEmail, user["email"])
	}
	if user["user_id"] == "" {
		t.Error("expected non-empty user_id")
	}
}

func TestMe_NoToken(t *testing.T) {
	srv := newTestServer(t)
	resp := doRequest(t, srv, http.MethodGet, "/auth/me", nil, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestMe_InvalidToken(t *testing.T) {
	srv := newTestServer(t)
	resp := doRequest(t, srv, http.MethodGet, "/auth/me", nil, "not.a.valid.token")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

// --- Logout ---

func TestLogout_Success(t *testing.T) {
	srv := newTestServer(t)
	accessToken, _ := registerAndLogin(t, srv)

	resp := doRequest(t, srv, http.MethodPost, "/auth/logout", nil, accessToken)
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}
}

func TestLogout_NoToken(t *testing.T) {
	srv := newTestServer(t)
	resp := doRequest(t, srv, http.MethodPost, "/auth/logout", nil, "")

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestLogout_RefreshInvalidatedAfterLogout(t *testing.T) {
	srv := newTestServer(t)
	accessToken, refreshToken := registerAndLogin(t, srv)

	// Logout
	logout := doRequest(t, srv, http.MethodPost, "/auth/logout", nil, accessToken)
	logout.Body.Close()

	// Attempt to use the refresh token — should now fail
	body := map[string]string{"email": testEmail, "refresh_token": refreshToken}
	resp := doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %v", r.Error)
	}
}

// --- Refresh ---

func TestRefresh_Success(t *testing.T) {
	srv := newTestServer(t)
	_, refreshToken := registerAndLogin(t, srv)

	body := map[string]string{"email": testEmail, "refresh_token": refreshToken}
	resp := doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	if !r.Success {
		t.Error("expected success: true")
	}

	var tokens map[string]string
	if err := json.Unmarshal(r.Data, &tokens); err != nil {
		t.Fatalf("unmarshal tokens: %v", err)
	}
	if tokens["access_token"] == "" {
		t.Error("expected non-empty access_token")
	}
	if tokens["refresh_token"] == "" {
		t.Error("expected non-empty refresh_token")
	}
}

func TestRefresh_InvalidToken(t *testing.T) {
	srv := newTestServer(t)
	registerAndLogin(t, srv)

	body := map[string]string{"email": testEmail, "refresh_token": "invalidtoken"}
	resp := doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %v", r.Error)
	}
}

func TestRefresh_ReusedToken(t *testing.T) {
	srv := newTestServer(t)
	_, refreshToken := registerAndLogin(t, srv)

	body := map[string]string{"email": testEmail, "refresh_token": refreshToken}

	// First refresh — should succeed and rotate the token
	first := doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	first.Body.Close()

	// Second refresh with the same token — should fail
	resp := doRequest(t, srv, http.MethodPost, "/auth/refresh", body, "")
	r := decodeResponse(t, resp)

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 on reused token, got %d", resp.StatusCode)
	}
	if r.Error == nil || r.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected UNAUTHORIZED, got %v", r.Error)
	}
}
