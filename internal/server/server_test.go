// Copyright 2025 Patryk Rostkowski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func defaultConfig(tmpDir string) *config.Config {
	return &config.Config{
		ServerConfig: &config.ServerConfig{
			ShutdownTimeout: 1 * time.Second,
		},
		HTTPServer: &config.HTTPServerConfig{
			Port: 0,
		},
		GRPCListener: &config.GRPCServerListenerConfig{
			GRPCServerTcpListener: &config.GRPCServerTcpListener{
				Port: 0,
			},
		},
		StoreConfig: &config.StoreConfig{
			DumpEnabled:    true,
			MemoryDumpPath: filepath.Join(tmpDir, "store.gob.gz"),
		},
	}
}

func TestPersistAndReadMemoryStore(t *testing.T) {
	cfg := defaultConfig(t.TempDir())

	logger := DefaultLogger()

	s1 := NewServer(logger, cfg, DefaultPrometheusRegistry())
	ctx := context.Background()

	_, err := s1.Set(ctx, &cachev1alpha.SetRequest{Key: "foo", Value: []byte("bar")})
	require.NoError(t, err)

	_, err = s1.Set(ctx, &cachev1alpha.SetRequest{Key: "baz", Value: []byte("qux")})
	require.NoError(t, err)

	err = s1.PersistMemoryStore()
	require.NoError(t, err)

	s2 := NewServer(logger, cfg, DefaultPrometheusRegistry())
	err = s2.ReadPersistedMemoryStore()
	require.NoError(t, err)

	resp, err := s2.Get(ctx, &cachev1alpha.GetRequest{Key: "foo"})
	require.NoError(t, err)
	assert.Equal(t, []byte("bar"), resp.Value)

	resp, err = s2.Get(ctx, &cachev1alpha.GetRequest{Key: "baz"})
	require.NoError(t, err)
	assert.Equal(t, []byte("qux"), resp.Value)
}

func TestReadPersistedMemoryStore_FileNotFound(t *testing.T) {
	s := NewTestServer(t)

	err := s.ReadPersistedMemoryStore()
	assert.NoError(t, err)
}

func TestReadPersistedMemoryStore_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "store.gob.gz")

	err := os.WriteFile(filePath, []byte{}, 0o600)
	require.NoError(t, err)

	cfg := defaultConfig(tmpDir)

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	err = s.ReadPersistedMemoryStore()
	assert.NoError(t, err)
}

func TestServerLifecycle_InitAndShutdown(t *testing.T) {
	cfg := defaultConfig("/tmp")

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())

	err := s.Init()
	assert.NoError(t, err)

	err = s.Shutdown()
	assert.NoError(t, err)
}

func TestServerStart_CancelContextTriggersShutdown(t *testing.T) {
	cfg := defaultConfig(t.TempDir())

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	require.NoError(t, s.Init())

	t.Cleanup(func() {
		err := s.Shutdown()
		require.NoError(t, err)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- s.Start(ctx)
	}()

	ready := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(ready)
	}()
	<-ready

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Server did not shut down in time")
	}
}

func TestServerInit_ReadPersistedMemoryStoreFails(t *testing.T) {
	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "bad-store.gob.gz")

	// Write corrupted file
	require.NoError(t, os.WriteFile(dumpPath, []byte("not a valid gob"), 0o600))

	cfg := defaultConfig(dumpPath)

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	err := s.Init()
	assert.Error(t, err)
}

func generateTestCertAndKey(certPath, keyPath string) error {
	// Generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	// Create certificate template
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "localhost",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// Create cert
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return err
	}

	// Write cert
	certOut, err := os.Create(certPath)
	if err != nil {
		return err
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: certDER}); err != nil {
		return err
	}

	// Write key
	keyOut, err := os.Create(keyPath)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	privBytes := x509.MarshalPKCS1PrivateKey(priv)
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: privBytes}); err != nil {
		return err
	}

	return nil
}

func TestGRPCServerFailsWithInvalidTLS(t *testing.T) {
	cfg := defaultConfig(t.TempDir())
	cfg.TLSConfig = &config.TLSConfig{
		Enabled:  true,
		CertFile: "/invalid/cert.pem",
		KeyFile:  "/invalid/key.pem",
	}

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	err := s.Init()
	assert.ErrorContains(t, err, "failed to load TLS credentials")
}

func TestHTTPServerTLS_StartsTLS(t *testing.T) {
	tmpDir := t.TempDir()
	certPath := filepath.Join(tmpDir, "cert.pem")
	keyPath := filepath.Join(tmpDir, "key.pem")

	err := generateTestCertAndKey(certPath, keyPath)
	require.NoError(t, err)

	cfg := defaultConfig(tmpDir)
	cfg.TLSConfig = &config.TLSConfig{
		Enabled:  true,
		CertFile: certPath,
		KeyFile:  keyPath,
	}

	cfg.HTTPServer.Port = 0

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	require.NoError(t, s.Init())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = s.Start(ctx)
	}()

	time.Sleep(200 * time.Millisecond) // give server time to start
	cancel()

	// Let shutdown finish
	time.Sleep(200 * time.Millisecond)
}

func TestServerInit_ListenerCreationFails(t *testing.T) {
	cfg := defaultConfig("/root") // assuming no permission

	// Invalid address/port
	cfg.GRPCListener.Address = ""
	cfg.GRPCListener.Port = -1

	s := NewServer(DefaultLogger(), cfg, DefaultPrometheusRegistry())
	err := s.Init()
	assert.Error(t, err)
}
