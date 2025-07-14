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

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, 50051, cfg.GRPCPort)
	assert.Equal(t, 9091, cfg.HTTPPort)
	assert.Equal(t, "0.0.0.0", cfg.ListenAddr)
	assert.False(t, cfg.DumpEnabled)
	assert.Equal(t, MemoryDumpPath, cfg.MemoryDumpPath)
	assert.Equal(t, MemoryDumpFileName, cfg.MemoryDumpFileName)
}

func TestMemoryDumpFileFullPath(t *testing.T) {
	cfg := DefaultConfig()
	expected := filepath.Join(cfg.MemoryDumpPath, cfg.MemoryDumpFileName)
	assert.Equal(t, expected, cfg.MemoryDumpFileFullPath())
}

func TestHTTPAndGRPCListenAddr(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "0.0.0.0:9091", cfg.HTTPListenAddr())
	assert.Equal(t, "0.0.0.0:50051", cfg.GRPCListenAddr())
}

func TestLoadConfig_Success(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "yaml")

	yaml := `
server:
  grpc_port: 1234
  http_port: 5678
  listen_addr: "127.0.0.1"
  shutdown_timeout: 20s
  graceful_timeout: 5s
store:
  dump_enabled: true
  memory_dump_path: "/tmp/cache/"
  memory_dump_file_name: "dump.gob.gz"
`
	err := os.WriteFile(yamlPath, []byte(yaml), 0600)
	require.NoError(t, err)

	originalPath := configFileFullPath
	configFileFullPath = func() string { return yamlPath }
	defer func() { configFileFullPath = originalPath }()

	cfg, err := LoadConfig()
	require.NoError(t, err)
	assert.Equal(t, 1234, cfg.GRPCPort)
	assert.Equal(t, "127.0.0.1", cfg.ListenAddr)
}

func TestLoadConfig_FileNotFound(t *testing.T) {
	originalPath := configFileFullPath
	configFileFullPath = func() string {
		return filepath.Join(t.TempDir(), "nonexistent.yaml")
	}
	defer func() { configFileFullPath = originalPath }()

	_, err := LoadConfig()
	assert.Error(t, err)
}

func TestLoadConfig_PartialFileDefaultsApplied(t *testing.T) {
	tmpDir := t.TempDir()
	configFileFullPath = func() string {
		return filepath.Join(tmpDir, "yaml")
	}

	partialYAML := `
server:
  grpc_port: 12345
store:
  dump_enabled: true
`

	err := os.WriteFile(configFileFullPath(), []byte(partialYAML), 0600)
	require.NoError(t, err)

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, 12345, cfg.GRPCPort)
	assert.True(t, cfg.DumpEnabled)

	assert.Equal(t, 9091, cfg.HTTPPort)
	assert.Equal(t, "0.0.0.0", cfg.ListenAddr)
	assert.Equal(t, ServerShutdownTimeout, cfg.ShutdownTimeout)
	assert.Equal(t, GracefulTimeout, cfg.GracefulTimeout)
	assert.Equal(t, MemoryDumpPath, cfg.MemoryDumpPath)
	assert.Equal(t, MemoryDumpFileName, cfg.MemoryDumpFileName)
}
