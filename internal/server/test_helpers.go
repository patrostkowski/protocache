package server

import (
	"path/filepath"
	"testing"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

func DefaultPrometheusRegistry() *prometheus.Registry {
	return prometheus.NewRegistry()
}

func NewTestServer(t *testing.T) *Server {
	t.Helper()

	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "store.gob.gz")

	cfg := &config.Config{
		StoreConfig: &v1alpha.StoreConfig{
			Engine:         v1alpha.MapStoreEngine,
			MemoryDumpPath: dumpPath,
		},
	}

	reg := DefaultPrometheusRegistry()

	return NewServer(cfg, reg)
}
