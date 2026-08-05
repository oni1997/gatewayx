package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"hash"
	"net/http"
	"strings"
	"time"
)

const NameHMAC = "hmac"

type HMACAuthenticator struct {
	secret     []byte
	algorithm  string
	headerName string
	clockSkew  time.Duration
}

type HMACOptions struct {
	Secret     string
	Algorithm  string
	HeaderName string
	ClockSkew  time.Duration
}

func NewHMAC(opts HMACOptions) (*HMACAuthenticator, error) {
	ha := &HMACAuthenticator{
		secret:     []byte(opts.Secret),
		algorithm:  opts.Algorithm,
		headerName: opts.HeaderName,
		clockSkew:  opts.ClockSkew,
	}

	if len(ha.secret) == 0 {
		return nil, fmt.Errorf("hmac: secret is required")
	}
	if ha.algorithm == "" {
		ha.algorithm = "sha256"
	}
	if ha.headerName == "" {
		ha.headerName = "X-Signature"
	}
	if ha.clockSkew == 0 {
		ha.clockSkew = 5 * time.Minute
	}

	return ha, nil
}

func (ha *HMACAuthenticator) Name() string {
	return NameHMAC
}

func (ha *HMACAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	sigHeader := r.Header.Get(ha.headerName)
	if sigHeader == "" {
		return nil, fmt.Errorf("missing signature header: %s", ha.headerName)
	}

	parts := strings.SplitN(sigHeader, "|", 3)
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid signature format, expected key_id|timestamp|signature")
	}

	keyID := parts[0]
	timestampStr := parts[1]
	signature := parts[2]

	ts, err := parseTimestamp(timestampStr)
	if err != nil {
		return nil, fmt.Errorf("invalid timestamp in signature: %w", err)
	}

	now := time.Now().UTC()
	diff := now.Sub(ts)
	if diff < 0 {
		diff = -diff
	}
	if diff > ha.clockSkew {
		return nil, fmt.Errorf("signature timestamp too far from server time")
	}

	var h hash.Hash
	switch ha.algorithm {
	case "sha256":
		h = hmac.New(sha256.New, ha.secret)
	case "sha512":
		h = hmac.New(sha512.New, ha.secret)
	default:
		return nil, fmt.Errorf("unsupported hmac algorithm: %s", ha.algorithm)
	}

	signingString := buildSigningString(r.Method, r.URL.Path, timestampStr, keyID)
	h.Write([]byte(signingString))
	expectedSig := hex.EncodeToString(h.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
		return nil, fmt.Errorf("invalid signature")
	}

	return Claims{
		"sub":    keyID,
		"type":   "hmac",
		"key_id": keyID,
		"method": ha.algorithm,
	}, nil
}

func buildSigningString(method, path, timestamp, keyID string) string {
	return fmt.Sprintf("%s\n%s\n%s\n%s", method, path, timestamp, keyID)
}

func parseTimestamp(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
	}
	for _, f := range formats {
		if ts, err := time.Parse(f, s); err == nil {
			return ts, nil
		}
	}
	var unixNano int64
	if _, err := fmt.Sscanf(s, "%d", &unixNano); err == nil {
		return time.Unix(0, unixNano), nil
	}
	return time.Time{}, fmt.Errorf("unable to parse timestamp: %s", s)
}
