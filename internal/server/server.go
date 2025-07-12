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
	"compress/gzip"
	"context"
	"encoding/gob"
	"errors"
	"log/slog"
	"os"
	"runtime"
	"sync"
	"time"

	pb "github.com/patrostkowski/protocache/api/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	MemoryDumpPath     = "/var/lib/protocache/"
	MemoryDumpFileName = "protocache.gob.gz"
)

var MemoryDumpFileFullPath = MemoryDumpPath + MemoryDumpFileName

type Server struct {
	pb.UnimplementedCacheServiceServer

	mu    sync.RWMutex
	store map[string][]byte

	logger *slog.Logger
}

func NewServer(logger *slog.Logger) *Server {
	return &Server{
		store:  make(map[string][]byte),
		logger: logger,
	}
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[req.Key] = req.Value
	return &pb.SetResponse{Success: true, Message: "OK"}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[req.Key]

	if !ok {
		CacheMisses.Inc()
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
	}

	CacheHits.Inc()
	return &pb.GetResponse{Found: true, Message: "found", Value: val}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.store[req.Key]; !ok {
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
	}

	delete(s.store, req.Key)
	return &pb.DeleteResponse{Success: true, Message: "deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string][]byte)
	return &pb.ClearResponse{Success: true, Message: "cleared"}, nil
}

func (s *Server) Stats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalBytes uint64
	for _, v := range s.store {
		totalBytes += uint64(len(v))
	}

	return &pb.StatsResponse{
		KeyCount:         uint64(len(s.store)),
		MemoryUsageBytes: totalBytes,
		GoVersion:        runtime.Version(),
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}

func (s *Server) PersistMemoryStore() error {
	if err := os.MkdirAll(MemoryDumpPath, 0700); err != nil {
		s.logger.Error("Failed to create directory for memory store dump", "error", err.Error())
		return err
	}

	f, err := os.OpenFile(MemoryDumpFileFullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		s.logger.Error("Failed to open memory store dump file", "error", err.Error())
		return err
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()

	e := gob.NewEncoder(gz)
	if err := e.Encode(s.store); err != nil {
		s.logger.Error("Failed to encode the memory store", "error", err.Error())
		return err
	}

	s.logger.Info("Succesfully written memory store dump to file", "path", MemoryDumpFileFullPath)
	return nil
}

func (s *Server) ReadPersistedMemoryStore() error {
	f, err := os.Open(MemoryDumpFileFullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.logger.Warn("Memory store dump file does not exist, starting with empty store")
			return nil
		}
		s.logger.Error("Failed to open memory store dump file", "error", err.Error())
		return err
	}
	defer f.Close()

	fs, err := f.Stat()
	if err != nil {
		s.logger.Error("Failed to get memory store dump file stats", "error", err.Error())
		return err
	}
	if fs.Size() == 0 {
		s.logger.Warn("Memory store dump is empty, skipping load")
		return nil
	}

	gz, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gz.Close()

	d := gob.NewDecoder(gz)
	if err := d.Decode(&s.store); err != nil {
		s.logger.Error("Failed to decode memory store dump file", "error", err.Error())
		return err
	}

	s.logger.Info("Succesfully read memory store dump to memory", "size", len(s.store))
	return nil
}
