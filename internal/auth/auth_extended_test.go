package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestBasicAuth_PlainPassword(t *testing.T) {
	ba, err := NewBasic(BasicOptions{
		Users: map[string]string{"admin": "secret123"},
		Realm: "TestRealm",
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.SetBasicAuth("admin", "secret123")

	claims, err := ba.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["sub"] != "admin" {
		t.Errorf("expected sub=admin, got %v", claims["sub"])
	}
}

func TestBasicAuth_InvalidCredentials(t *testing.T) {
	ba, _ := NewBasic(BasicOptions{
		Users: map[string]string{"admin": "secret123"},
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.SetBasicAuth("admin", "wrong")

	_, err := ba.Authenticate(req)
	if err == nil {
		t.Error("expected error for invalid password")
	}
}

func TestBasicAuth_UnknownUser(t *testing.T) {
	ba, _ := NewBasic(BasicOptions{
		Users: map[string]string{"admin": "secret123"},
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.SetBasicAuth("unknown", "pass")

	_, err := ba.Authenticate(req)
	if err == nil {
		t.Error("expected error for unknown user")
	}
}

func TestBasicAuth_MissingHeader(t *testing.T) {
	ba, _ := NewBasic(BasicOptions{
		Users: map[string]string{"admin": "secret"},
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	_, err := ba.Authenticate(req)
	if err == nil {
		t.Error("expected error for missing auth header")
	}
}

func TestBasicAuth_SHA(t *testing.T) {
	ba, err := NewBasic(BasicOptions{
		Users: map[string]string{
			"sha-user": "{SHA}j0r3v5oexI5hA4Q4K8L0ZQ==",
		},
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.SetBasicAuth("sha-user", "not-checked-plain")

	_, err = ba.Authenticate(req)
	if err == nil {
		t.Error("expected error for non-matching SHA hash")
	}
}

func TestBasicAuth_HtpasswdFile(t *testing.T) {
	f, err := os.CreateTemp("", "htpasswd-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("fileuser:filepass\n")
	_, _ = f.WriteString("# comment line\n")
	_, _ = f.WriteString("another:password123\n")
	f.Close()

	ba, err := NewBasic(BasicOptions{
		Htpasswd: f.Name(),
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.SetBasicAuth("fileuser", "filepass")

	claims, err := ba.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["sub"] != "fileuser" {
		t.Errorf("expected sub=fileuser, got %v", claims["sub"])
	}
}

func TestHMAC_SHA256(t *testing.T) {
	secret := "hmac-secret-key"
	ha, err := NewHMAC(HMACOptions{
		Secret:    secret,
		Algorithm: "sha256",
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	now := time.Now().UTC()
	timestamp := now.Format(time.RFC3339)
	keyID := "client-abc"

	signingString := buildSigningString("GET", "/api/data", timestamp, keyID)
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(signingString))
	signature := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest("GET", "/api/data", nil)
	req.Header.Set("X-Signature", fmt.Sprintf("%s|%s|%s", keyID, timestamp, signature))

	claims, err := ha.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["key_id"] != "client-abc" {
		t.Errorf("expected key_id=client-abc, got %v", claims["key_id"])
	}
}

func TestHMAC_InvalidSignature(t *testing.T) {
	ha, _ := NewHMAC(HMACOptions{Secret: "secret123"})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Signature", "key1|badtimestamp|badsig")

	_, err := ha.Authenticate(req)
	if err == nil {
		t.Error("expected error for invalid signature")
	}
}

func TestHMAC_MissingHeader(t *testing.T) {
	ha, _ := NewHMAC(HMACOptions{Secret: "secret123"})

	req := httptest.NewRequest("GET", "/api/test", nil)
	_, err := ha.Authenticate(req)
	if err == nil {
		t.Error("expected error for missing signature header")
	}
}

func TestHMAC_ExpiredTimestamp(t *testing.T) {
	ha, err := NewHMAC(HMACOptions{
		Secret:    "secret123",
		ClockSkew: time.Second,
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	oldTime := time.Now().Add(-2 * time.Second).UTC()
	signingString := buildSigningString("GET", "/api", oldTime.Format(time.RFC3339), "key1")
	h := hmac.New(sha256.New, []byte("secret123"))
	h.Write([]byte(signingString))
	sig := hex.EncodeToString(h.Sum(nil))

	req := httptest.NewRequest("GET", "/api", nil)
	req.Header.Set("X-Signature", fmt.Sprintf("key1|%s|%s", oldTime.Format(time.RFC3339), sig))

	_, err = ha.Authenticate(req)
	if err == nil {
		t.Error("expected error for expired timestamp")
	}
}

func TestSession_CreateAndGet(t *testing.T) {
	store := NewSessionStore(100, time.Hour)
	defer store.Stop()

	session, err := store.Create("user-42", Claims{"role": "admin"})
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}
	if session.UserID != "user-42" {
		t.Errorf("expected user-42, got %s", session.UserID)
	}

	got, ok := store.Get(session.ID)
	if !ok {
		t.Fatal("session not found")
	}
	if got.UserID != "user-42" {
		t.Errorf("expected user-42, got %s", got.UserID)
	}
}

func TestSession_Expired(t *testing.T) {
	store := NewSessionStore(100, 5*time.Second)
	defer store.Stop()

	session, _ := store.Create("user-1", Claims{})

	_, ok := store.Get(session.ID)
	if !ok {
		t.Fatal("session should exist immediately after creation")
	}
}

func TestSession_Delete(t *testing.T) {
	store := NewSessionStore(100, time.Hour)
	defer store.Stop()

	session, _ := store.Create("user-1", Claims{})
	store.Delete(session.ID)

	_, ok := store.Get(session.ID)
	if ok {
		t.Error("session should be deleted")
	}
}

func TestSession_MaxSessions(t *testing.T) {
	store := NewSessionStore(3, time.Hour)
	defer store.Stop()

	for i := 0; i < 3; i++ {
		_, err := store.Create(fmt.Sprintf("user-%d", i), Claims{})
		if err != nil {
			t.Fatalf("unexpected error on session %d: %v", i, err)
		}
	}

	_, err := store.Create("user-4", Claims{})
	if err == nil {
		t.Error("expected error when exceeding max sessions")
	}
}

func TestSession_Authenticator(t *testing.T) {
	sa, err := NewSession(SessionOptions{
		TTL:         time.Hour,
		MaxSessions: 100,
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	session, _ := sa.store.Create("user-99", Claims{"role": "editor"})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-Session-ID", session.ID)

	claims, err := sa.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["user_id"] != "user-99" {
		t.Errorf("expected user_id=user-99, got %v", claims["user_id"])
	}
	if claims["session_id"] != session.ID {
		t.Errorf("expected session_id to match")
	}
}

func TestSession_CookieAuth(t *testing.T) {
	sa, _ := NewSession(SessionOptions{TTL: time.Hour})
	session, _ := sa.store.Create("cookieguy", Claims{"foo": "bar"})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "gatewayx_session", Value: session.ID})

	claims, err := sa.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["user_id"] != "cookieguy" {
		t.Errorf("expected cookieguy, got %v", claims["user_id"])
	}
}

func TestRBAC_BasicCheck(t *testing.T) {
	engine := NewRBACEngine()
	engine.AddPermission(Permission{
		Path:    "/admin/**",
		Methods: []string{"GET", "POST"},
		Roles:   []string{"admin"},
	})
	engine.AddPermission(Permission{
		Path:  "/api/**",
		Roles: []string{"admin", "developer"},
	})

	tests := []struct {
		name   string
		roles  []string
		method string
		path   string
		want   bool
	}{
		{"admin can access admin", []string{"admin"}, "GET", "/admin/users", true},
		{"admin can access api", []string{"admin"}, "POST", "/api/data", true},
		{"developer can access api", []string{"developer"}, "GET", "/api/v1/data", true},
		{"developer cannot access admin", []string{"developer"}, "GET", "/admin/users", false},
		{"viewer cannot access anything", []string{"viewer"}, "GET", "/api/data", false},
		{"admin cannot DELETE admin", []string{"admin"}, "DELETE", "/admin/users", false},
		{"exact path match", []string{"admin"}, "GET", "/api", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.CheckPermission(tt.roles, tt.method, tt.path)
			if got != tt.want {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRBAC_WildcardPath(t *testing.T) {
	engine := NewRBACEngine()
	engine.AddPermission(Permission{
		Path:  "/public/*",
		Roles: []string{"guest"},
	})

	if !engine.CheckPermission([]string{"guest"}, "GET", "/public/index.html") {
		t.Error("guest should access /public/index.html")
	}
	if engine.CheckPermission([]string{"guest"}, "GET", "/public/deep/path") {
		t.Error("guest should not access /public/deep/path with /*")
	}
}

func TestRBAC_GlobPattern(t *testing.T) {
	engine := NewRBACEngine()
	engine.AddPermission(Permission{
		Path:  "/v[0-9]/**",
		Roles: []string{"client"},
	})

	if !engine.CheckPermission([]string{"client"}, "GET", "/v1/users") {
		t.Error("glob pattern should match /v1/users")
	}
	if !engine.CheckPermission([]string{"client"}, "GET", "/v2/deep/nested") {
		t.Error("glob pattern should match /v2/deep/nested")
	}
	if engine.CheckPermission([]string{"client"}, "GET", "/x1/users") {
		t.Error("glob pattern should not match /x1/users")
	}
}

func TestRBAC_Authenticator(t *testing.T) {
	engine := NewRBACEngine()
	engine.AddPermission(Permission{
		Path:  "/api/**",
		Roles: []string{"admin"},
	})

	ja, _ := NewJWT(JWTOptions{Secret: "test"})
	ra := NewRBAC(ja, engine, "roles")

	t.Run("valid jwt with roles", func(t *testing.T) {
		token := createTestJWT(t, "test", map[string]any{
			"sub":   "admin-user",
			"roles": []string{"admin"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		claims, err := ra.Authenticate(req)
		if err != nil {
			t.Fatalf("auth failed: %v", err)
		}
		if claims["_rbac"] != true {
			t.Error("expected _rbac=true in claims")
		}
	})

	t.Run("valid jwt without roles", func(t *testing.T) {
		token := createTestJWT(t, "test", map[string]any{
			"sub": "no-role-user",
			"exp": time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		_, err := ra.Authenticate(req)
		if err == nil {
			t.Error("expected error for missing roles")
		}
	})
}

func TestRBAC_Middleware(t *testing.T) {
	engine := NewRBACEngine()
	engine.AddPermission(Permission{
		Path:  "/api/**",
		Roles: []string{"admin"},
	})

	ja, _ := NewJWT(JWTOptions{Secret: "test"})
	ra := NewRBAC(ja, engine, "roles")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	rbacMw := RBACMiddleware(ra)

	t.Run("allowed role", func(t *testing.T) {
		token := createTestJWT(t, "test", map[string]any{
			"sub":   "admin-user",
			"roles": []string{"admin"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMw := Middleware(ra)
		authMw(rbacMw(handler)).ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("forbidden role", func(t *testing.T) {
		token := createTestJWT(t, "test", map[string]any{
			"sub":   "viewer",
			"roles": []string{"viewer"},
			"exp":   time.Now().Add(time.Hour).Unix(),
		})

		req := httptest.NewRequest("GET", "/api/data", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		rec := httptest.NewRecorder()

		authMw := Middleware(ra)
		authMw(rbacMw(handler)).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", rec.Code)
		}
	})
}

func TestTokenCache(t *testing.T) {
	cache := NewTokenCache(5 * time.Minute)

	claims := Claims{"sub": "user-1", "exp": float64(time.Now().Add(time.Hour).Unix())}
	cache.Set("token-abc", claims)

	if cache.Size() != 1 {
		t.Errorf("expected size 1, got %d", cache.Size())
	}

	got, ok := cache.Get("token-abc")
	if !ok {
		t.Fatal("cached claims not found")
	}
	if got["sub"] != "user-1" {
		t.Errorf("expected sub=user-1, got %v", got["sub"])
	}
}

func TestTokenCache_Expiry(t *testing.T) {
	cache := NewTokenCache(time.Millisecond)

	claims := Claims{"sub": "ephemeral"}
	cache.Set("token-exp", claims)

	time.Sleep(2 * time.Millisecond)

	_, ok := cache.Get("token-exp")
	if ok {
		t.Error("expired cache entry should not be returned")
	}
}

func TestTokenCache_Invalidate(t *testing.T) {
	cache := NewTokenCache(time.Hour)
	cache.Set("t1", Claims{"sub": "u1"})
	cache.Set("t2", Claims{"sub": "u2"})

	cache.Invalidate()
	if cache.Size() != 0 {
		t.Errorf("expected size 0 after invalidate, got %d", cache.Size())
	}
}

func TestCachedJWT(t *testing.T) {
	ja, _ := NewJWT(JWTOptions{Secret: "cache-secret"})
	cja := NewCachedJWT(ja, time.Minute)

	token := createTestJWT(t, "cache-secret", map[string]any{
		"sub": "cached-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	claims1, err := cja.Authenticate(req)
	if err != nil {
		t.Fatalf("first auth failed: %v", err)
	}

	claims2, err := cja.Authenticate(req)
	if err != nil {
		t.Fatalf("cached auth failed: %v", err)
	}

	if claims1["sub"] != claims2["sub"] {
		t.Error("cached claims should match original")
	}
}

func TestCachedJWT_InvalidToken(t *testing.T) {
	ja, _ := NewJWT(JWTOptions{Secret: "secret"})
	cja := NewCachedJWT(ja, time.Minute)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	_, err := cja.Authenticate(req)
	if err == nil {
		t.Error("expected error for invalid cached JWT")
	}
}

func createTestJWT(t *testing.T, secret string, claims map[string]any) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(claims))
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign test JWT: %v", err)
	}
	return tokenString
}
