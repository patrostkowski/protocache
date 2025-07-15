package server

import (
	"context"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	transport "github.com/Jille/raft-grpc-transport"
	"github.com/hashicorp/raft"
	boltdb "github.com/hashicorp/raft-boltdb"
	internalconfig "github.com/patrostkowski/protocache/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *Server) raftListenAddr() string {
	ip, err := internalconfig.GetContainerIP()
	if err != nil {
		panic(err)
	}
	return net.JoinHostPort(ip, strconv.Itoa(s.config.ServerConfig.GRPCPort))
}

func (s *Server) InitRaft(ctx context.Context) error {
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(s.config.ID)

	baseDir := filepath.Join(internalconfig.RaftDirPath, s.config.ID)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return fmt.Errorf("failed to create raft base directory %q: %w", baseDir, err)
	}

	logStore, err := boltdb.NewBoltStore(filepath.Join(baseDir, "logs.dat"))
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}

	stableStore, err := boltdb.NewBoltStore(filepath.Join(baseDir, "stable.dat"))
	if err != nil {
		return fmt.Errorf("failed to create stable store: %w", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(baseDir, 3, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}

	s.transportManager = transport.New(
		raft.ServerAddress(s.raftListenAddr()),
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())},
	)

	s.raft, err = raft.NewRaft(config, s, logStore, stableStore, snapshotStore, s.transportManager.Transport())
	if err != nil {
		return fmt.Errorf("failed to start Raft: %w", err)
	}

	var servers []raft.Server
	for _, addr := range s.config.ClusterMembers {
		servers = append(servers, raft.Server{
			ID:       raft.ServerID(strings.Split(addr, ":")[0]),
			Address:  raft.ServerAddress(addr),
			Suffrage: raft.Voter,
		})
	}

	// Only bootstrap if this node is the first in the list
	if s.config.InitCluster {
		hasState, err := raft.HasExistingState(logStore, stableStore, snapshotStore)
		if err != nil {
			return fmt.Errorf("failed to check existing raft state: %w", err)
		}
		if !hasState {
			s.logger.Info("Bootstrapping initial Raft cluster", "servers", servers)
			future := s.raft.BootstrapCluster(raft.Configuration{Servers: servers})
			if err := future.Error(); err != nil && err != raft.ErrCantBootstrap {
				return fmt.Errorf("raft bootstrap failed: %w", err)
			}
		}
	}

	return nil
}

func (s Store) Persist(sink raft.SnapshotSink) error {
	defer sink.Close()

	encoder := gob.NewEncoder(sink)
	if err := encoder.Encode(s); err != nil {
		sink.Cancel()
		return err
	}
	return nil
}

func (s Store) Release() {}

func (s *Server) Apply(log *raft.Log) interface{} {
	var cmd struct {
		Op    string
		Key   string
		Value []byte
	}

	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		s.logger.Error("Failed to unmarshal raft log", "error", err)
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch cmd.Op {
	case "set":
		s.store[cmd.Key] = cmd.Value
	case "delete":
		delete(s.store, cmd.Key)
	default:
		return fmt.Errorf("unknown command: %s", cmd.Op)
	}

	return nil
}

func (s *Server) Snapshot() (raft.FSMSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Deep copy to avoid mutation during snapshot
	snapshot := make(Store, len(s.store))
	for k, v := range s.store {
		snapshot[k] = append([]byte(nil), v...)
	}
	return snapshot, nil
}

func (s *Server) Restore(rc io.ReadCloser) error {
	defer rc.Close()

	var restored Store
	decoder := gob.NewDecoder(rc)
	if err := decoder.Decode(&restored); err != nil {
		return err
	}

	s.mu.Lock()
	s.store = restored
	s.mu.Unlock()
	return nil
}
