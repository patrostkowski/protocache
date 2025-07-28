package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	// GRPC TCP listener default
	assert.NotNil(t, cfg.GRPCListener)
	assert.NotNil(t, cfg.GRPCListener.GRPCServerTcpListener)
	assert.Equal(t, 50051, cfg.GRPCListener.Port)
	assert.Equal(t, "0.0.0.0", cfg.GRPCListener.Address)

	// HTTP server default
	assert.NotNil(t, cfg.HTTPServer)
	assert.Equal(t, 9091, cfg.HTTPServer.Port)
	assert.Equal(t, "0.0.0.0:9091", cfg.HTTPListenAddr())

	// Store config
	assert.NotNil(t, cfg.StoreConfig)
	assert.False(t, cfg.StoreConfig.DumpEnabled)
	assert.Equal(t, MemoryDumpPath, cfg.StoreConfig.MemoryDumpPath)
	assert.Equal(t, MemoryDumpFileName, cfg.StoreConfig.MemoryDumpFileName)
}

func TestMemoryDumpFileFullPath(t *testing.T) {
	cfg := DefaultConfig()
	expected := filepath.Join(cfg.StoreConfig.MemoryDumpPath, cfg.StoreConfig.MemoryDumpFileName)
	assert.Equal(t, expected, cfg.MemoryDumpFileFullPath())
}

func TestHTTPAndGRPCListenAddr(t *testing.T) {
	cfg := DefaultConfig()
	assert.Equal(t, "0.0.0.0:9091", cfg.HTTPListenAddr())
	assert.Equal(t, "0.0.0.0:50051", cfg.GRPCListenAddr())
}

func TestLoadConfig_TCP_Success(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "tcp_config.yaml")

	yaml := `
grpc_listener:
  address: "127.0.0.1"
  port: 1234

http_server:
  address: "127.0.0.1"
  port: 5678

store:
  dump_enabled: true
  memory_dump_path: "/tmp/cache/"
  memory_dump_file_name: "dump.gob.gz"
`

	err := os.WriteFile(yamlPath, []byte(yaml), 0o600)
	require.NoError(t, err)

	originalPath := configFileFullPath
	configFileFullPath = func() string { return yamlPath }
	defer func() { configFileFullPath = originalPath }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "127.0.0.1", cfg.GRPCListener.Address)
	assert.Equal(t, 1234, cfg.GRPCListener.Port)

	assert.Equal(t, "127.0.0.1", cfg.HTTPServer.Address)
	assert.Equal(t, 5678, cfg.HTTPServer.Port)

	assert.True(t, cfg.StoreConfig.DumpEnabled)
	assert.Equal(t, "/tmp/cache/", cfg.StoreConfig.MemoryDumpPath)
	assert.Equal(t, "dump.gob.gz", cfg.StoreConfig.MemoryDumpFileName)
}

func TestLoadConfig_Unix_Success(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "unix_config.yaml")

	yaml := `
grpc_listener:
  socket_path: "/tmp/grpc.sock"

http_server:
  address: "localhost"
  port: 8080
`

	err := os.WriteFile(yamlPath, []byte(yaml), 0o600)
	require.NoError(t, err)

	originalPath := configFileFullPath
	configFileFullPath = func() string { return yamlPath }
	defer func() { configFileFullPath = originalPath }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	assert.Equal(t, "/tmp/grpc.sock", cfg.GRPCListener.SocketPath)
	assert.Equal(t, "localhost", cfg.HTTPServer.Address)
	assert.Equal(t, 8080, cfg.HTTPServer.Port)
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

func TestLoadConfig_PartialDefaultsApplied(t *testing.T) {
	tmpDir := t.TempDir()
	yamlPath := filepath.Join(tmpDir, "partial.yaml")

	yaml := `
grpc_listener:
  address: "localhost"
  port: 1111

store:
  dump_enabled: true
`

	err := os.WriteFile(yamlPath, []byte(yaml), 0o600)
	require.NoError(t, err)

	originalPath := configFileFullPath
	configFileFullPath = func() string { return yamlPath }
	defer func() { configFileFullPath = originalPath }()

	cfg, err := LoadConfig()
	require.NoError(t, err)

	// GRPC
	assert.Equal(t, "localhost", cfg.GRPCListener.Address)
	assert.Equal(t, 1111, cfg.GRPCListener.Port)

	// Defaults still applied
	assert.NotNil(t, cfg.HTTPServer)
	assert.Equal(t, HTTPPort, cfg.HTTPServer.Port)
	assert.Equal(t, ListenAddr, cfg.HTTPServer.Address)

	assert.True(t, cfg.StoreConfig.DumpEnabled)
	assert.Equal(t, MemoryDumpPath, cfg.StoreConfig.MemoryDumpPath)
	assert.Equal(t, MemoryDumpFileName, cfg.StoreConfig.MemoryDumpFileName)
}

func TestCreateListener_TCP(t *testing.T) {
	cfg := DefaultConfig()

	// Use a random port to avoid conflicts
	cfg.GRPCListener.Port = 0

	lis, err := cfg.CreateListener()
	require.NoError(t, err)
	require.NotNil(t, lis)

	addr := lis.Addr().String()
	assert.Contains(t, addr, ":")

	err = lis.Close()
	assert.NoError(t, err)
}

func TestCreateListener_Unix(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := filepath.Join(tmpDir, "test.sock")

	cfg := &Config{
		GRPCListener: &v1alpha.GRPCServerListenerConfig{
			GRPCServerUnixListener: &v1alpha.GRPCServerUnixListener{
				SocketPath: socketPath,
			},
		},
	}

	lis, err := cfg.CreateListener()
	require.NoError(t, err)
	require.NotNil(t, lis)

	// Confirm the socket file exists
	_, err = os.Stat(socketPath)
	assert.NoError(t, err, "expected socket file to be created")

	err = lis.Close()
	assert.NoError(t, err)
}

func TestCreateListener_Unix_MissingSocketPath(t *testing.T) {
	cfg := &Config{
		GRPCListener: &v1alpha.GRPCServerListenerConfig{
			GRPCServerUnixListener: &v1alpha.GRPCServerUnixListener{
				SocketPath: "",
			},
		},
	}

	_, err := cfg.CreateListener()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unix socket path is empty")
}

func TestCreateListener_MissingGRPCListener(t *testing.T) {
	cfg := &Config{
		GRPCListener: nil,
	}

	_, err := cfg.CreateListener()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GRPCListener config is nil")
}
