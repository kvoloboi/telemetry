package tlsconfig

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

type Config struct {
	Enabled            bool
	CACertPath         string
	CertPath           string
	KeyPath            string
	InsecureSkipVerify bool
	ServerName         string
}

func (c Config) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.CACertPath == "" || c.CertPath == "" || c.KeyPath == "" {
		return fmt.Errorf("tls enabled but cert paths are not fully set")
	}
	return nil
}

func loadCertPool(caPath string) (*x509.CertPool, error) {
	caBytes, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("read ca cert: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caBytes) {
		return nil, fmt.Errorf("failed to append CA cert")
	}

	return pool, nil
}

func ClientTLSConfig(cfg Config) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load client cert: %w", err)
	}

	caPool, err := loadCertPool(cfg.CACertPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caPool,
		ServerName:         cfg.ServerName,
		InsecureSkipVerify: cfg.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS13,
	}, nil
}

func ServerTLSConfig(cfg Config) (*tls.Config, error) {
	if !cfg.Enabled {
		return nil, nil
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("load server cert: %w", err)
	}

	caPool, err := loadCertPool(cfg.CACertPath)
	if err != nil {
		return nil, err
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
