package auth

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestJWT_HS256(t *testing.T) {
	secret := "test-secret-key"
	ja, err := NewJWT(JWTOptions{Secret: secret, Algorithm: "HS256"})
	if err != nil {
		t.Fatalf("failed to create JWT auth: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	claims, err := ja.Authenticate(req)
	if err != nil {
		t.Fatalf("authenticate failed: %v", err)
	}

	if claims["sub"] != "user-123" {
		t.Errorf("expected sub=user-123, got %v", claims["sub"])
	}
}

func TestJWT_Expired(t *testing.T) {
	secret := "test-secret"
	ja, _ := NewJWT(JWTOptions{Secret: secret})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte(secret))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	_, err := ja.Authenticate(req)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestJWT_MissingHeader(t *testing.T) {
	ja, _ := NewJWT(JWTOptions{Secret: "test"})
	req := httptest.NewRequest("GET", "/api/test", nil)

	_, err := ja.Authenticate(req)
	if err == nil {
		t.Error("expected error for missing header")
	}
}

func TestJWT_SecretFile(t *testing.T) {
	f, err := os.CreateTemp("", "jwt-secret-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("file-secret-key")
	f.Close()

	ja, err := NewJWT(JWTOptions{SecretFile: f.Name()})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "file-user",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("file-secret-key"))

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	claims, err := ja.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["sub"] != "file-user" {
		t.Errorf("expected sub=file-user, got %v", claims["sub"])
	}
}

func TestAPIKey_Header(t *testing.T) {
	ak, err := NewAPIKey(APIKeyOptions{
		Keys:       map[string]string{"key-abc": "user-1"},
		HeaderName: "X-API-Key",
	})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "key-abc")

	claims, err := ak.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["sub"] != "user-1" {
		t.Errorf("expected sub=user-1, got %v", claims["sub"])
	}
}

func TestAPIKey_InvalidKey(t *testing.T) {
	ak, _ := NewAPIKey(APIKeyOptions{
		Keys: map[string]string{"valid-key": "user-1"},
	})

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "bad-key")

	_, err := ak.Authenticate(req)
	if err == nil {
		t.Error("expected error for invalid key")
	}
}

func TestAPIKey_MissingHeader(t *testing.T) {
	ak, _ := NewAPIKey(APIKeyOptions{
		Keys: map[string]string{"key-1": "user-1"},
	})

	req := httptest.NewRequest("GET", "/api/test", nil)

	_, err := ak.Authenticate(req)
	if err == nil {
		t.Error("expected error for missing header")
	}
}

func TestAPIKey_KeysFile(t *testing.T) {
	f, err := os.CreateTemp("", "apikeys-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	_, _ = f.WriteString("file-key-xyz: file-user\n")
	f.Close()

	ak, err := NewAPIKey(APIKeyOptions{KeysFile: f.Name()})
	if err != nil {
		t.Fatalf("failed: %v", err)
	}

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("X-API-Key", "file-key-xyz")

	claims, err := ak.Authenticate(req)
	if err != nil {
		t.Fatalf("auth failed: %v", err)
	}
	if claims["sub"] != "file-user" {
		t.Errorf("expected sub=file-user, got %v", claims["sub"])
	}
}

func TestRegistry(t *testing.T) {
	reg := NewRegistry()
	ak, _ := NewAPIKey(APIKeyOptions{Keys: map[string]string{"k": "u"}})
	reg.Register(ak)

	a, ok := reg.Get("api_key")
	if !ok {
		t.Fatal("expected authenticator in registry")
	}
	if a.Name() != "api_key" {
		t.Errorf("expected api_key, got %s", a.Name())
	}
}

func TestWithClaims(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	claims := Claims{"sub": "test-user"}
	ctx := WithClaims(req.Context(), claims)

	got, ok := GetClaims(ctx)
	if !ok {
		t.Fatal("expected claims in context")
	}
	if got["sub"] != "test-user" {
		t.Errorf("expected sub=test-user, got %v", got["sub"])
	}
}

func TestMiddlewareBlocksUnauthenticated(t *testing.T) {
	ja, _ := NewJWT(JWTOptions{Secret: "test"})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Middleware(ja)(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMiddlewarePassesAuthenticated(t *testing.T) {
	ja, _ := NewJWT(JWTOptions{Secret: "test"})

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": "user-1",
		"exp": time.Now().Add(time.Hour).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("test"))

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := GetClaims(r.Context())
		w.Header().Set("X-User", claims["sub"].(string))
		w.WriteHeader(http.StatusOK)
	})

	wrapped := Middleware(ja)(handler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-User") != "user-1" {
		t.Errorf("expected X-User=user-1, got %s", rec.Header().Get("X-User"))
	}
}
