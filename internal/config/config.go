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
	"log/slog"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/yaml.v3"
)

const (
	GRPCPort                  = 50051
	HTTPPort                  = 9091
	ListenAddr                = "0.0.0.0"
	UnixSocketPath            = "/var/run/protocache/protocache.sock"
	ServerShutdownTimeout     = 30 * time.Second
	GracefulTimeout           = 10 * time.Second
	MemoryDumpPath            = "/var/lib/protocache/"
	MemoryDumpFileName        = "protocache.gob.gz"
	ConfigFilePath            = "/etc/protocache/"
	ConfigFileName            = "config.yaml"
	ConfigFileDefaultFilePath = ConfigFilePath + ConfigFileName
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
	GRPCListener *v1alpha.GRPCServerListenerConfig `yaml:"grpc_listener"`
	HTTPServer   *v1alpha.HTTPServerConfig         `yaml:"http_server"`
	ServerConfig *v1alpha.ServerConfig             `yaml:"server"`
	StoreConfig  *v1alpha.StoreConfig              `yaml:"store"`
	TLSConfig    *v1alpha.TLSConfig                `yaml:"tls"`
}

func DefaultConfig() *Config {
	return &Config{
		GRPCListener: &v1alpha.GRPCServerListenerConfig{
			GRPCServerTcpListener: &v1alpha.GRPCServerTcpListener{
				Address: ListenAddr,
				Port:    GRPCPort,
			},
		},
		HTTPServer: &v1alpha.HTTPServerConfig{
			Address: ListenAddr,
			Port:    HTTPPort,
		},
		StoreConfig: &v1alpha.StoreConfig{
			Engine:             v1alpha.MapStoreEngine,
			DumpEnabled:        false,
			MemoryDumpPath:     MemoryDumpPath,
			MemoryDumpFileName: MemoryDumpFileName,
		},
		TLSConfig: &v1alpha.TLSConfig{
			Enabled: false,
		},
	}
}

func (c *Config) GRPCListenerType() v1alpha.GRPCServerListenerType {
	if c.GRPCListener.SocketPath != "" {
		return v1alpha.UNIX
	}
	return v1alpha.TCP
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

func LoadConfig(path string) (*Config, error) {
	logger.Info("Attempting to load configuration", slog.String("path", path))

	data, err := os.ReadFile(path)
	if err != nil {
		logger.Warn("Could not read config file", slog.String("path", path), slog.Any("error", err))
		return nil, err
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		logger.Error("Failed to parse configuration file", slog.Any("error", err))
		return nil, err
	}

	applyDefaults(cfg)
	logger.Info("Configuration loaded successfully")
	return cfg, nil
}

func applyDefaults(cfg *Config) {
	defaults := DefaultConfig()

	if cfg.GRPCListener == nil {
		cfg.GRPCListener = defaults.GRPCListener
	}
	if cfg.HTTPServer == nil {
		cfg.HTTPServer = defaults.HTTPServer
	}
	if cfg.StoreConfig == nil {
		cfg.StoreConfig = defaults.StoreConfig
	}
	if cfg.GRPCListener.GRPCServerTcpListener == nil && defaults.GRPCListener.GRPCServerTcpListener != nil {
		cfg.GRPCListener.GRPCServerTcpListener = defaults.GRPCListener.GRPCServerTcpListener
	}
	if cfg.GRPCListener.GRPCServerUnixListener == nil && defaults.GRPCListener.GRPCServerUnixListener != nil {
		cfg.GRPCListener.GRPCServerUnixListener = defaults.GRPCListener.GRPCServerUnixListener
	}
	if cfg.HTTPServer.Address == "" {
		cfg.HTTPServer.Address = defaults.HTTPServer.Address
	}
	if cfg.HTTPServer.Port == 0 {
		cfg.HTTPServer.Port = defaults.HTTPServer.Port
	}
	if cfg.StoreConfig.MemoryDumpPath == "" {
		cfg.StoreConfig.MemoryDumpPath = defaults.StoreConfig.MemoryDumpPath
	}
	if cfg.StoreConfig.MemoryDumpFileName == "" {
		cfg.StoreConfig.MemoryDumpFileName = defaults.StoreConfig.MemoryDumpFileName
	}
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
	logger.Info("Initializing TLS credentials", slog.String("cert", c.TLSConfig.CertFile), slog.String("key", c.TLSConfig.KeyFile))

	if c.TLSConfig == nil || !c.TLSConfig.Enabled {
		return nil, nil
	}

	creds, err := credentials.NewServerTLSFromFile(c.TLSConfig.CertFile, c.TLSConfig.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS credentials: %w", err)
	}

	return grpc.Creds(creds), nil
}

func (c *Config) GetStoreEngine() v1alpha.StoreEngine {
	return c.StoreConfig.Engine
}

func (c *Config) GetEvictionPolicy() v1alpha.EvictionPolicy {
	return c.StoreConfig.EvictionPolicy
}
