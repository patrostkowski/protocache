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
	"context"
	"log/slog"
	"maps"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"slices"
	"sync"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/config"
)

type Server struct {
	pb.UnimplementedCacheServiceServer

	mu    sync.RWMutex
	store map[string][]byte

	logger     *slog.Logger
	config     *config.Config
	listener   *net.Listener
	grpcServer *grpc.Server
	httpServer *http.Server
	metrics    *grpcprom.ServerMetrics
	registry   prometheus.Registerer
}

func NewServer(logger *slog.Logger, config *config.Config, reg prometheus.Registerer) *Server {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	return &Server{
		store:    make(map[string][]byte),
		logger:   logger,
		config:   config,
		registry: reg,
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

func (s *Server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := slices.Collect(maps.Keys(s.store))
	return &pb.ListResponse{Keys: keys}, nil
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

func (s *Server) Init() error {
	s.metrics = grpcprom.NewServerMetrics()
	if err := s.registry.Register(s.metrics); err != nil {
		return err
	}

	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.metrics.UnaryServerInterceptor(),
			LoggingUnaryInterceptor(s.logger),
		),
		grpc.ConnectionTimeout(s.config.ServerConfig.ShutdownTimeout),
	)

	pb.RegisterCacheServiceServer(s.grpcServer, s)
	s.metrics.InitializeMetrics(s.grpcServer)
	reflection.Register(s.grpcServer)

	listener, err := s.config.CreateListener()
	if err != nil {
		return err
	}
	s.listener = &listener

	s.httpServer = &http.Server{
		Addr:    s.config.HTTPListenAddr(),
		Handler: s.metricsHandler(),
	}

	if s.config.IsMemoryStoreDumpEnabled() {
		if err := s.ReadPersistedMemoryStore(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		s.logger.Info("Starting HTTP server", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	go func() {
		s.logger.Info("Starting gRPC server", "addr", (*s.listener).Addr().String())
		if err := s.grpcServer.Serve(*s.listener); err != nil {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		s.logger.Info("Shutdown signal received")
	case err := <-errCh:
		s.logger.Error("Server crashed", "error", err)
		return err
	}

	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	shutdownTimer := time.AfterFunc(s.config.ServerConfig.GracefulTimeout, func() {
		s.logger.Warn("Graceful shutdown timeout exceeded, forcing stop")
		s.grpcServer.Stop()
	})
	defer shutdownTimer.Stop()

	s.grpcServer.GracefulStop()

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		s.logger.Error("Error shutting down HTTP server", "error", err)
	}

	if s.config.IsMemoryStoreDumpEnabled() {
		if err := s.PersistMemoryStore(); err != nil {
			s.logger.Error("Failed to persist memory store", "error", err)
			return err
		}
	}

	s.logger.Info("Server shut down cleanly")
	return nil
}

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config from file, using default config", "error", err)
		cfg = config.DefaultConfig()
	}

	srv := NewServer(logger, cfg, prometheus.DefaultRegisterer)

	if err := srv.Init(); err != nil {
		logger.Error("Initialization failed", "error", err)
		return err
	}

	if err := srv.Start(ctx); err != nil {
		return err
	}

	return nil
}
