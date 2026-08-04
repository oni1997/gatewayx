package auth

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const NameJWT = "jwt"

type JWTAuthenticator struct {
	secretKey     []byte
	publicKey     any
	privateKey    any
	algorithm     string
	claimsHeaders map[string]string
}

type JWTOptions struct {
	Secret        string
	SecretFile    string
	PublicKeyFile string
	Algorithm     string
	HeaderName    string
	ClaimsHeader  map[string]string
}

func NewJWT(opts JWTOptions) (*JWTAuthenticator, error) {
	ja := &JWTAuthenticator{
		algorithm:     opts.Algorithm,
		claimsHeaders: opts.ClaimsHeader,
	}

	if ja.algorithm == "" {
		ja.algorithm = "HS256"
	}

	headerName := opts.HeaderName
	if headerName == "" {
		headerName = "Authorization"
	}
	_ = headerName

	switch {
	case opts.Secret != "":
		ja.secretKey = []byte(opts.Secret)
	case opts.SecretFile != "":
		data, err := os.ReadFile(opts.SecretFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file: %w", err)
		}
		ja.secretKey = []byte(strings.TrimSpace(string(data)))
	case opts.PublicKeyFile != "":
		data, err := os.ReadFile(opts.PublicKeyFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read public key file: %w", err)
		}
		rsaKey, rsaErr := jwt.ParseRSAPublicKeyFromPEM(data)
		if rsaErr == nil {
			ja.publicKey = rsaKey
		} else {
			ecKey, ecErr := jwt.ParseECPublicKeyFromPEM(data)
			if ecErr != nil {
				return nil, fmt.Errorf("failed to parse public key (rsa: %w, ecdsa: %w)", rsaErr, ecErr)
			}
			ja.publicKey = ecKey
		}
	default:
		return nil, fmt.Errorf("jwt: one of secret, secret_file, or public_key_file must be provided")
	}

	return ja, nil
}

func (ja *JWTAuthenticator) Name() string {
	return NameJWT
}

func (ja *JWTAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	tokenString := extractBearerToken(r)
	if tokenString == "" {
		return nil, fmt.Errorf("missing or malformed authorization header")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{ja.algorithm}),
		jwt.WithLeeway(30*time.Second),
	)

	var keyFunc jwt.Keyfunc
	switch {
	case ja.secretKey != nil:
		keyFunc = func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return ja.secretKey, nil
		}
	case ja.publicKey != nil:
		keyFunc = func(t *jwt.Token) (any, error) {
			return ja.publicKey, nil
		}
	default:
		return nil, fmt.Errorf("jwt: no key configured")
	}

	token, err := parser.Parse(tokenString, keyFunc)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims := make(Claims)
	if mapClaims, ok := token.Claims.(jwt.MapClaims); ok {
		for k, v := range mapClaims {
			claims[k] = v
		}
	}

	return claims, nil
}

func extractBearerToken(r *http.Request) string {
	auth := r.Header.Get("Authorization")
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}
