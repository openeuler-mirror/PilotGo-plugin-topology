package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

// SetupTLSLoadCA creates an X509 certificate from file
func SetupTLSLoadCA(certPEM string) (*x509.CertPool, error) {
	rootPEM, err := ioutil.ReadFile(certPEM)
	if err != nil {
		return nil, fmt.Errorf("Failed to open root certificate '%s' : %s", certPEM, err)
	}
	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootPEM))
	if !ok {
		return nil, fmt.Errorf("Failed to parse root certificate '%s'", rootPEM)
	}
	return roots, nil
}

// SetupTLSClientConfig creates a client X509 certificate from public and private key
func SetupTLSClientConfig(certPEM string, keyPEM string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, fmt.Errorf("Can't read X509 key pair set in config : cert '%s' key '%s' : %s", certPEM, keyPEM, err)
	}
	cfgTLS := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return cfgTLS, nil
}

// SetupTLSServerConfig creates a server X509 certificate from public and private key
func SetupTLSServerConfig(certPEM string, keyPEM string) (*tls.Config, error) {
	cfgTLS, err := SetupTLSClientConfig(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	cfgTLS.MinVersion = tls.VersionTLS12
	cfgTLS.ClientAuth = tls.VerifyClientCertIfGiven
	cfgTLS.CurvePreferences = []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256}
	cfgTLS.PreferServerCipherSuites = true
	cfgTLS.CipherSuites = []uint16{
		tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
	}

	return cfgTLS, nil
}
