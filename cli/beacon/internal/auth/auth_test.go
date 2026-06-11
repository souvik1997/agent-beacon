package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGeneratePKCECreatesVerifierChallengeAndState(t *testing.T) {
	params, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE returned error: %v", err)
	}
	base64URL := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	for name, value := range map[string]string{
		"code verifier":  params.CodeVerifier,
		"code challenge": params.CodeChallenge,
		"state":          params.State,
	} {
		if len(value) != 43 {
			t.Fatalf("%s length = %d, want 43", name, len(value))
		}
		if !base64URL.MatchString(value) {
			t.Fatalf("%s contains non-base64url characters: %q", name, value)
		}
	}
	hash := sha256.Sum256([]byte(params.CodeVerifier))
	wantChallenge := base64.RawURLEncoding.EncodeToString(hash[:])
	if params.CodeChallenge != wantChallenge {
		t.Fatalf("CodeChallenge = %q, want %q", params.CodeChallenge, wantChallenge)
	}
}

func TestSaveLoadCredentialsPrivatePermissions(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	creds := &Credentials{
		Token:       "secret-token",
		TokenPrefix: "asym_123",
		UserID:      "user-1",
		Email:       "user@example.test",
		OrgName:     "Example Org",
	}
	if err := SaveCredentials(creds); err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}
	path, err := CredentialsPath()
	if err != nil {
		t.Fatalf("CredentialsPath returned error: %v", err)
	}
	if got, want := path, filepath.Join(home, ".beacon", "auth", "credentials.json"); got != want {
		t.Fatalf("CredentialsPath = %q, want %q", got, want)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat credentials dir: %v", err)
	}
	if got, want := dirInfo.Mode().Perm(), os.FileMode(0700); got != want {
		t.Fatalf("credentials dir permissions = %o, want %o", got, want)
	}
	fileInfo, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat credentials file: %v", err)
	}
	if got, want := fileInfo.Mode().Perm(), os.FileMode(0600); got != want {
		t.Fatalf("credentials permissions = %o, want %o", got, want)
	}

	loaded, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials returned error: %v", err)
	}
	if loaded.Token != "secret-token" || loaded.Email != "user@example.test" {
		t.Fatalf("credentials did not round-trip: %#v", loaded)
	}
	if !IsLoggedIn() {
		t.Fatal("IsLoggedIn = false, want true")
	}
}

func TestIsLoggedInRejectsExpiredCredentials(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	if err := SaveCredentials(&Credentials{
		Token:     "expired-token",
		UserID:    "user-1",
		ExpiresAt: time.Now().Add(-time.Minute),
	}); err != nil {
		t.Fatalf("SaveCredentials returned error: %v", err)
	}
	if IsLoggedIn() {
		t.Fatal("IsLoggedIn = true, want false for expired credentials")
	}
}

func TestCallbackServerExchangesCodeForToken(t *testing.T) {
	var gotRequest exchangeCodeRequest
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/cli/auth/exchange" {
			t.Fatalf("exchange path = %q, want /api/cli/auth/exchange", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("exchange method = %q, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
			t.Fatalf("decode exchange request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"token":"secret-token","token_prefix":"asym_123","user_id":"user-1","email":"user@example.test","org_id":"org-1","org_name":"Example Org"}`))
	}))
	defer backend.Close()

	server, err := NewCallbackServer("state-123", "verifier-123", backend.URL, backend.Client())
	if err != nil {
		t.Fatalf("NewCallbackServer returned error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()
	server.Start()

	callbackURL := url.URL{
		Scheme:   "http",
		Host:     "127.0.0.1:" + strconv.Itoa(server.Port()),
		Path:     "/callback",
		RawQuery: "exchange_code=code-123&state=state-123",
	}
	go func() {
		resp, err := http.Get(callbackURL.String())
		if err == nil {
			_ = resp.Body.Close()
		}
	}()

	result, err := server.Wait(2 * time.Second)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if result.Error != "" {
		t.Fatalf("callback result error = %q", result.Error)
	}
	if result.Token != "secret-token" || result.Email != "user@example.test" || result.OrgName != "Example Org" {
		t.Fatalf("unexpected callback result: %#v", result)
	}
	if gotRequest.ExchangeCode != "code-123" || gotRequest.State != "state-123" || gotRequest.CodeVerifier != "verifier-123" {
		t.Fatalf("unexpected exchange request: %#v", gotRequest)
	}
}

func TestCallbackServerRejectsStateMismatch(t *testing.T) {
	backendCalled := false
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		backendCalled = true
		t.Fatal("backend should not be called on state mismatch")
	}))
	defer backend.Close()

	server, err := NewCallbackServer("expected-state", "verifier", backend.URL, backend.Client())
	if err != nil {
		t.Fatalf("NewCallbackServer returned error: %v", err)
	}
	defer func() { _ = server.Shutdown() }()
	server.Start()

	go func() {
		resp, err := http.Get("http://127.0.0.1:" + strconv.Itoa(server.Port()) + "/callback?exchange_code=code&state=wrong-state")
		if err == nil {
			_ = resp.Body.Close()
		}
	}()
	result, err := server.Wait(2 * time.Second)
	if err != nil {
		t.Fatalf("Wait returned error: %v", err)
	}
	if !strings.Contains(result.Error, "state mismatch") {
		t.Fatalf("result error = %q, want state mismatch", result.Error)
	}
	if backendCalled {
		t.Fatal("backend was called on state mismatch")
	}
}

func TestResolveDashboardURL(t *testing.T) {
	t.Setenv(DashboardURLEnv, "")
	if got := ResolveDashboardURL(""); got != DefaultDashboardURL {
		t.Fatalf("ResolveDashboardURL empty = %q, want default", got)
	}
	t.Setenv(DashboardURLEnv, "https://env.example/")
	if got := ResolveDashboardURL(""); got != "https://env.example" {
		t.Fatalf("ResolveDashboardURL env = %q, want env without trailing slash", got)
	}
	if got := ResolveDashboardURL("https://flag.example///"); got != "https://flag.example" {
		t.Fatalf("ResolveDashboardURL flag = %q, want flag without trailing slash", got)
	}
}
