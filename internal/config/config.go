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

	"gopkg.in/yaml.v3"
)

const (
	GRPCPort              = 50051
	HTTPPort              = 9091
	ListenAddr            = "0.0.0.0"
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

type ServerConfig struct {
	GRPCPort        int           `yaml:"grpc_port"`
	HTTPPort        int           `yaml:"http_port"`
	ListenAddr      string        `yaml:"listen_addr"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
	GracefulTimeout time.Duration `yaml:"graceful_timeout"`
}

type StoreConfig struct {
	DumpEnabled        bool   `yaml:"dump_enabled"`
	MemoryDumpPath     string `yaml:"memory_dump_path"`
	MemoryDumpFileName string `yaml:"memory_dump_file_name"`
}

type Config struct {
	ServerConfig `yaml:"server"`
	StoreConfig  `yaml:"store"`
}

func DefaultConfig() *Config {
	return &Config{
		ServerConfig: ServerConfig{
			GRPCPort:        GRPCPort,
			HTTPPort:        HTTPPort,
			ListenAddr:      ListenAddr,
			ShutdownTimeout: ServerShutdownTimeout,
			GracefulTimeout: GracefulTimeout,
		},
		StoreConfig: StoreConfig{
			DumpEnabled:        false,
			MemoryDumpPath:     MemoryDumpPath,
			MemoryDumpFileName: MemoryDumpFileName,
		},
	}
}

func (c *Config) MemoryDumpFileFullPath() string {
	return filepath.Join(c.StoreConfig.MemoryDumpPath, c.StoreConfig.MemoryDumpFileName)
}

func (c *Config) HTTPListenAddr() string {
	return net.JoinHostPort(c.ServerConfig.ListenAddr, strconv.Itoa(c.ServerConfig.HTTPPort))
}

func (c *Config) GRPCListenAddr() string {
	return net.JoinHostPort(c.ServerConfig.ListenAddr, strconv.Itoa(c.ServerConfig.GRPCPort))
}

func (c *Config) IsMemoryStoreDumpEnabled() bool {
	return c.StoreConfig.DumpEnabled
}

func LoadConfig() (*Config, error) {
	cfg := DefaultConfig() // ← use defaults
	data, err := os.ReadFile(configFileFullPath())
	if err != nil {
		return nil, err
	}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
