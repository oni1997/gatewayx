package oauth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oni1997/gatewayx/internal/auth"
)

func testFlow(t *testing.T) *Flow {
	t.Helper()
	a, err := auth.NewOAuth(auth.OAuthOptions{
		Provider:     "github",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})
	if err != nil {
		t.Fatalf("failed to create oauth authenticator: %v", err)
	}
	return NewFlow(a)
}

func TestLoginHandler_SetsStateCookieAndRedirects(t *testing.T) {
	f := testFlow(t)

	req := httptest.NewRequest("GET", "/oauth/login", nil)
	rec := httptest.NewRecorder()

	f.LoginHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected state cookie to be set")
	}
	if cookies[0].Name != "gatewayx_oauth_state" {
		t.Errorf("expected gatewayx_oauth_state cookie, got %s", cookies[0].Name)
	}
	if cookies[0].Value == "" {
		t.Error("state cookie should have a value")
	}

	location := rec.Header().Get("Location")
	if location == "" {
		t.Error("expected redirect location")
	}
}

func TestCallbackHandler_MissingState(t *testing.T) {
	f := testFlow(t)

	req := httptest.NewRequest("GET", "/oauth/callback?code=abc", nil)
	rec := httptest.NewRecorder()

	f.CallbackHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCallbackHandler_MissingCode(t *testing.T) {
	f := testFlow(t)

	req := httptest.NewRequest("GET", "/oauth/callback?state=xyz", nil)
	rec := httptest.NewRecorder()

	f.CallbackHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCallbackHandler_InvalidState(t *testing.T) {
	f := testFlow(t)

	req := httptest.NewRequest("GET", "/oauth/callback?state=wrong&code=abc", nil)
	req.AddCookie(&http.Cookie{Name: "gatewayx_oauth_state", Value: "correct"})
	rec := httptest.NewRecorder()

	f.CallbackHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for state mismatch, got %d", rec.Code)
	}
}

func TestLogoutHandler_ClearsCookieAndRedirects(t *testing.T) {
	f := testFlow(t)

	req := httptest.NewRequest("GET", "/oauth/logout", nil)
	rec := httptest.NewRecorder()

	f.LogoutHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusFound {
		t.Errorf("expected 302, got %d", rec.Code)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "gatewayx_oauth" && c.Value == "" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected gatewayx_oauth cookie to be cleared")
	}
}

func TestRandomState(t *testing.T) {
	state1 := randomState()
	state2 := randomState()

	if state1 == "" {
		t.Error("state should not be empty")
	}
	if len(state1) != 32 {
		t.Errorf("expected 32 char hex state, got %d", len(state1))
	}
	if state1 == state2 {
		t.Error("two random states should differ")
	}
}
