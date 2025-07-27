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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/patrostkowski/protocache/internal/store"
	"github.com/patrostkowski/protocache/internal/store/mapstore"
	"github.com/patrostkowski/protocache/internal/store/syncmapstore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
)

const (
	GRPCPort              = 50051
	HTTPPort              = 9091
	ListenAddr            = "0.0.0.0"
	UnixSocketPath        = "/var/run/protocache/protocache.sock"
	ServerShutdownTimeout = 30 * time.Second
	GracefulTimeout       = 10 * time.Second
	MemoryDumpPath        = "/var/lib/protocache/"
	MemoryDumpFileName    = "protocache.gob.gz"
	ConfigFilePath        = "/etc/protocache/"
	ConfigFileName        = "config.yaml"
)

var (
	GRPCAddr = fmt.Sprintf("%s:%d", ListenAddr, GRPCPort)
	HTTPAddr = fmt.Sprintf("%s:%d", ListenAddr, HTTPPort)

	MemoryDumpFileFullPath = MemoryDumpPath + MemoryDumpFileName
)

var configFileFullPath = func() string {
	return filepath.Join(ConfigFilePath, ConfigFileName)
}

type Config struct {
	GRPCListener *GRPCServerListenerConfig `yaml:"grpc_listener"`
	HTTPServer   *HTTPServerConfig         `yaml:"http_server"`
	ServerConfig *ServerConfig             `yaml:"server"`
	StoreConfig  *StoreConfig              `yaml:"store"`
	TLSConfig    *TLSConfig                `yaml:"tls"`
}

type GRPCServerListenerType int

const (
	UNIX GRPCServerListenerType = iota
	TCP
)

type GRPCServerListenerConfig struct {
	*GRPCServerUnixListener `yaml:",inline"`
	*GRPCServerTcpListener  `yaml:",inline"`
}

type GRPCServerUnixListener struct {
	SocketPath string `yaml:"socket_path"`
}

type GRPCServerTcpListener struct {
	Address string `yaml:"address"` // ":50051" or "/tmp/protocache.sock"
	Port    int    `yaml:"port"`
}

type HTTPServerConfig struct {
	Address string `yaml:"address"` // e.g., "0.0.0.0:9091"
	Port    int    `yaml:"port"`
}

type ServerConfig struct {
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	GracefulTimeout time.Duration `yaml:"graceful_timeout"`
}

type StoreEngine string

const (
	MapStoreEngine     StoreEngine = "map"
	SyncMapStoreEngine StoreEngine = "syncmap"
)

type StoreConfig struct {
	Engine             StoreEngine `yaml:"engine"`
	DumpEnabled        bool        `yaml:"dump_enabled"`
	MemoryDumpPath     string      `yaml:"memory_dump_path"`
	MemoryDumpFileName string      `yaml:"memory_dump_file_name"`
}

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func DefaultConfig() *Config {
	return &Config{
		GRPCListener: &GRPCServerListenerConfig{
			GRPCServerTcpListener: &GRPCServerTcpListener{
				Address: ListenAddr,
				Port:    GRPCPort,
			},
		},
		HTTPServer: &HTTPServerConfig{
			Address: ListenAddr,
			Port:    HTTPPort,
		},
		StoreConfig: &StoreConfig{
			Engine:             MapStoreEngine,
			DumpEnabled:        false,
			MemoryDumpPath:     MemoryDumpPath,
			MemoryDumpFileName: MemoryDumpFileName,
		},
		TLSConfig: &TLSConfig{
			Enabled: false,
		},
	}
}

func (c *Config) GRPCListenerType() GRPCServerListenerType {
	if c.GRPCListener.SocketPath != "" {
		return UNIX
	}
	return TCP
}

func (c *Config) MemoryDumpFileFullPath() string {
	return filepath.Join(c.StoreConfig.MemoryDumpPath, c.StoreConfig.MemoryDumpFileName)
}

func (c *Config) HTTPListenAddr() string {
	return net.JoinHostPort(c.HTTPServer.Address, strconv.Itoa(c.HTTPServer.Port))
}

func (c *Config) GRPCListenAddr() string {
	return net.JoinHostPort(c.GRPCListener.Address, strconv.Itoa(c.GRPCListener.Port))
}

func (c *Config) IsMemoryStoreDumpEnabled() bool {
	return c.StoreConfig.DumpEnabled
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(configFileFullPath())
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	// Apply default fallbacks manually
	defaults := DefaultConfig()

	// Fill missing nested structs
	if cfg.GRPCListener == nil {
		cfg.GRPCListener = defaults.GRPCListener
	}
	if cfg.HTTPServer == nil {
		cfg.HTTPServer = defaults.HTTPServer
	}
	if cfg.StoreConfig == nil {
		cfg.StoreConfig = defaults.StoreConfig
	}

	// Fill TCP listener defaults if not overridden
	if cfg.GRPCListener.GRPCServerTcpListener == nil && defaults.GRPCListener.GRPCServerTcpListener != nil {
		cfg.GRPCListener.GRPCServerTcpListener = defaults.GRPCListener.GRPCServerTcpListener
	}

	// Fill Unix listener if provided via defaults (only needed if you set it by default)
	if cfg.GRPCListener.GRPCServerUnixListener == nil && defaults.GRPCListener.GRPCServerUnixListener != nil {
		cfg.GRPCListener.GRPCServerUnixListener = defaults.GRPCListener.GRPCServerUnixListener
	}

	// Fill HTTP port/address if missing
	if cfg.HTTPServer.Address == "" {
		cfg.HTTPServer.Address = defaults.HTTPServer.Address
	}
	if cfg.HTTPServer.Port == 0 {
		cfg.HTTPServer.Port = defaults.HTTPServer.Port
	}

	// Same for Store
	if cfg.StoreConfig.MemoryDumpPath == "" {
		cfg.StoreConfig.MemoryDumpPath = defaults.StoreConfig.MemoryDumpPath
	}
	if cfg.StoreConfig.MemoryDumpFileName == "" {
		cfg.StoreConfig.MemoryDumpFileName = defaults.StoreConfig.MemoryDumpFileName
	}

	return cfg, nil
}

func (c *Config) CreateListener() (net.Listener, error) {
	if c.GRPCListener == nil {
		return nil, fmt.Errorf("GRPCListener config is nil")
	}

	if c.GRPCListener.GRPCServerUnixListener != nil {
		socketPath := c.GRPCListener.SocketPath
		if socketPath == "" {
			return nil, fmt.Errorf("unix socket path is empty")
		}

		dir := filepath.Dir(socketPath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("failed to create socket directory: %w", err)
		}

		if _, err := os.Stat(socketPath); err == nil {
			if err := os.Remove(socketPath); err != nil {
				return nil, fmt.Errorf("failed to remove existing unix socket file: %w", err)
			}
		}

		return net.Listen("unix", socketPath)
	}

	return net.Listen(
		"tcp",
		fmt.Sprintf("%s:%d",
			c.GRPCListener.Address,
			c.GRPCListener.Port,
		))
}

func (c *Config) CreateTLS() (grpc.ServerOption, error) {
	if c.TLSConfig == nil || !c.TLSConfig.Enabled {
		return nil, nil
	}

	creds, err := credentials.NewServerTLSFromFile(c.TLSConfig.CertFile, c.TLSConfig.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	return grpc.Creds(creds), nil
}

func (c *Config) NewStore() (store.Store, error) {
	switch c.StoreConfig.Engine {
	case SyncMapStoreEngine:
		return syncmapstore.NewSyncMapStore(), nil
	default:
		return mapstore.NewMapStore(), nil
	}
}
