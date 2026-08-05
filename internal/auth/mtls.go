package auth

import (
	"crypto/x509"
	"fmt"
	"net/http"
	"os"
)

const NameMTLS = "mtls"

type mTLSAuthenticator struct {
	caCertPool  *x509.CertPool
	verifyDepth int
}

type MTLS struct {
	CACertFile  string
	VerifyDepth int
}

func NewMTLS(opts MTLS) (*mTLSAuthenticator, error) {
	ma := &mTLSAuthenticator{
		verifyDepth: opts.VerifyDepth,
	}

	if ma.verifyDepth == 0 {
		ma.verifyDepth = 1
	}

	data, err := os.ReadFile(opts.CACertFile)
	if err != nil {
		return nil, fmt.Errorf("mtls: failed to read CA cert: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("mtls: failed to parse CA certificate from %s", opts.CACertFile)
	}
	ma.caCertPool = pool

	return ma, nil
}

func (ma *mTLSAuthenticator) Name() string {
	return NameMTLS
}

func (ma *mTLSAuthenticator) Authenticate(r *http.Request) (Claims, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return nil, fmt.Errorf("mtls: no client certificate provided")
	}

	cert := r.TLS.PeerCertificates[0]

	opts := x509.VerifyOptions{
		Roots:         ma.caCertPool,
		Intermediates: x509.NewCertPool(),
	}

	if len(r.TLS.PeerCertificates) > 1 {
		for _, intermediate := range r.TLS.PeerCertificates[1:] {
			opts.Intermediates.AddCert(intermediate)
		}
	}

	if _, err := cert.Verify(opts); err != nil {
		return nil, fmt.Errorf("mtls: certificate verification failed: %w", err)
	}

	identity := cert.Subject.CommonName
	if identity == "" && len(cert.DNSNames) > 0 {
		identity = cert.DNSNames[0]
	}
	if identity == "" {
		identity = cert.SerialNumber.String()
	}

	return Claims{
		"sub":      identity,
		"type":     "mtls",
		"cn":       cert.Subject.CommonName,
		"serial":   cert.SerialNumber.String(),
		"not_before": cert.NotBefore,
		"not_after":  cert.NotAfter,
	}, nil
}
