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

package v1alpha

import "time"

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

type TLSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

type StoreEngine string

const (
	MapStoreEngine     StoreEngine = "map"
	SyncMapStoreEngine StoreEngine = "syncmap"
)

type EvictionPolicy string

const (
	EvictionNone   EvictionPolicy = "none"
	EvictionLRU    EvictionPolicy = "lru"
	EvictionLFU    EvictionPolicy = "lfu"
	EvictionRandom EvictionPolicy = "random"
)

type StoreConfig struct {
	Engine             StoreEngine    `yaml:"engine"`
	EvictionPolicy     EvictionPolicy `yaml:"eviction_policy"`
	DumpEnabled        bool           `yaml:"dump_enabled"`
	MemoryDumpPath     string         `yaml:"memory_dump_path"`
	MemoryDumpFileName string         `yaml:"memory_dump_file_name"`
}
