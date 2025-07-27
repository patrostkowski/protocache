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

type StoreConfig struct {
	Engine             StoreEngine `yaml:"engine"`
	DumpEnabled        bool        `yaml:"dump_enabled"`
	MemoryDumpPath     string      `yaml:"memory_dump_path"`
	MemoryDumpFileName string      `yaml:"memory_dump_file_name"`
}
