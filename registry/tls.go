package registry

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

// tlsInfo holds TLS certificate information for creating tls.Config.
// This is a simple helper struct to avoid depending on etcd's transport package.
type tlsInfo struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

// newTLSInfo creates a tlsInfo from TLSConfig.
func newTLSInfo(cfg *TLSConfig) (*tlsInfo, error) {
	if cfg == nil || !cfg.Enabled {
		return nil, nil
	}

	if cfg.CertFile == "" {
		return nil, fmt.Errorf("TLS cert file is required when TLS is enabled")
	}
	if cfg.KeyFile == "" {
		return nil, fmt.Errorf("TLS key file is required when TLS is enabled")
	}
	if cfg.CAFile == "" {
		return nil, fmt.Errorf("TLS CA file is required when TLS is enabled")
	}

	return &tlsInfo{
		CertFile: cfg.CertFile,
		KeyFile:  cfg.KeyFile,
		CAFile:   cfg.CAFile,
	}, nil
}

// ClientConfig creates a tls.Config for client connections.
func (info *tlsInfo) ClientConfig() (*tls.Config, error) {
	if info == nil {
		return nil, nil
	}

	// Load client certificate
	cert, err := tls.LoadX509KeyPair(info.CertFile, info.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load CA certificate
	caData, err := os.ReadFile(info.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caData) {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}
