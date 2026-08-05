package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const NameOAuth = "oauth"

var supportedProviders = map[string]oauthProvider{
	"github": {
		authURL:     "https://github.com/login/oauth/authorize",
		tokenURL:    "https://github.com/login/oauth/access_token",
		userInfoURL: "https://api.github.com/user",
		emailURL:    "https://api.github.com/user/emails",
		idField:     "login",
		acceptHeader: "application/vnd.github.v3+json",
	},
	"google": {
		authURL:      "https://accounts.google.com/o/oauth2/v2/auth",
		tokenURL:     "https://oauth2.googleapis.com/token",
		userInfoURL:  "https://www.googleapis.com/oauth2/v3/userinfo",
		idField:      "email",
	},
}

type oauthProvider struct {
	authURL      string
	tokenURL     string
	userInfoURL  string
	emailURL     string
	idField      string
	acceptHeader string
}

type OAuthAuthenticator struct {
	clientID     string
	clientSecret string
	provider     oauthProvider
	redirects    map[string]string
	client       *http.Client
}

type OAuthOptions struct {
	Provider     string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

func NewOAuth(opts OAuthOptions) (*OAuthAuthenticator, error) {
	prov, ok := supportedProviders[opts.Provider]
	if !ok {
		return nil, fmt.Errorf("oauth: unsupported provider %s (supported: github, google)", opts.Provider)
	}

	if opts.ClientID == "" || opts.ClientSecret == "" {
		return nil, fmt.Errorf("oauth: client_id and client_secret are required")
	}

	return &OAuthAuthenticator{
		clientID:     opts.ClientID,
		clientSecret: opts.ClientSecret,
		provider:     prov,
		redirects:    make(map[string]string),
		client:       &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (oa *OAuthAuthenticator) Name() string {
	return NameOAuth
}

func (oa *OAuthAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	token := extractBearerToken(r)
	if token == "" {
		return nil, fmt.Errorf("missing OAuth token")
	}

	userInfo, err := oa.fetchUserInfo(token)
	if err != nil {
		return nil, fmt.Errorf("oauth token validation failed: %w", err)
	}

	id, _ := userInfo[oa.provider.idField].(string)
	if id == "" {
		return nil, fmt.Errorf("oauth: could not extract user identity")
	}

	claims := Claims{
		"sub":    id,
		"type":   "oauth",
		"provider": oa.provider.tokenURL,
	}
	for k, v := range userInfo {
		claims[k] = v
	}

	return claims, nil
}

func (oa *OAuthAuthenticator) fetchUserInfo(token string) (map[string]any, error) {
	req, err := http.NewRequest("GET", oa.provider.userInfoURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	if oa.provider.acceptHeader != "" {
		req.Header.Set("Accept", oa.provider.acceptHeader)
	}

	resp, err := oa.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("userinfo returned %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse userinfo: %w", err)
	}

	return result, nil
}

func (oa *OAuthAuthenticator) ExchangeCode(code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", oa.clientID)
	form.Set("client_secret", oa.clientSecret)
	form.Set("code", code)

	req, err := http.NewRequest("POST", oa.provider.tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if oa.provider.acceptHeader != "" {
		req.Header.Set("Accept", oa.provider.acceptHeader)
	}

	resp, err := oa.client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token exchange failed: %s", string(body))
	}

	parsed, err := url.ParseQuery(string(body))
	if err != nil {
		return "", err
	}

	token := parsed.Get("access_token")
	if token == "" {
		return "", fmt.Errorf("no access_token in response: %s", string(body))
	}

	return token, nil
}
