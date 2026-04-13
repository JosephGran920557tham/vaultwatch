package tls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	inttls "github.com/yourusername/vaultwatch/internal/tls"
)

func writeTempPEM(t *testing.T, pemData []byte) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "cert*.pem")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = f.Write(pemData)
	_ = f.Close()
	return f.Name()
}

func generateSelfSignedCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		t.Fatal(err)
	}
	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	keyBytes, _ := x509.MarshalECPrivateKey(priv)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	return
}

func TestBuild_Defaults(t *testing.T) {
	cfg := inttls.Config{}
	tlsCfg, err := inttls.Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify=false by default")
	}
}

func TestBuild_InsecureSkipVerify(t *testing.T) {
	cfg := inttls.Config{InsecureSkipVerify: true}
	tlsCfg, err := inttls.Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !tlsCfg.InsecureSkipVerify {
		t.Error("expected InsecureSkipVerify=true")
	}
}

func TestBuild_WithCACert(t *testing.T) {
	certPEM, _ := generateSelfSignedCert(t)
	path := writeTempPEM(t, certPEM)
	cfg := inttls.Config{CACertFile: path}
	tlsCfg, err := inttls.Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tlsCfg.RootCAs == nil {
		t.Error("expected non-nil RootCAs")
	}
}

func TestBuild_WithClientCert(t *testing.T) {
	certPEM, keyPEM := generateSelfSignedCert(t)
	certPath := writeTempPEM(t, certPEM)
	keyPath := writeTempPEM(t, keyPEM)
	cfg := inttls.Config{ClientCertFile: certPath, ClientKeyFile: keyPath}
	tlsCfg, err := inttls.Build(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tlsCfg.Certificates) != 1 {
		t.Errorf("expected 1 client certificate, got %d", len(tlsCfg.Certificates))
	}
}

func TestBuild_InvalidCACertFile(t *testing.T) {
	cfg := inttls.Config{CACertFile: "/nonexistent/ca.pem"}
	_, err := inttls.Build(cfg)
	if err == nil {
		t.Error("expected error for missing CA cert file")
	}
}

func TestBuild_InvalidPEM(t *testing.T) {
	path := writeTempPEM(t, []byte("not-valid-pem"))
	cfg := inttls.Config{CACertFile: path}
	_, err := inttls.Build(cfg)
	if err == nil {
		t.Error("expected error for invalid PEM data")
	}
}
