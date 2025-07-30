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
