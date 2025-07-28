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
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/store"
)

type Server struct {
	cachev1alpha.UnimplementedCacheServiceServer

	store      store.Store
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
	store := store.NewStore(config.GetStoreEngine())
	return &Server{
		store:    store,
		logger:   logger,
		config:   config,
		registry: reg,
		metrics:  grpcprom.NewServerMetrics(), // ✅ this must be here
	}
}

func (s *Server) initGRPCServer() error {
	opts := []grpc.ServerOption{}

	if tlsOpt, err := s.config.CreateTLS(); err != nil {
		return err
	} else if tlsOpt != nil {
		opts = append(opts, tlsOpt)
	}

	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			s.metrics.UnaryServerInterceptor(),
			LoggingUnaryInterceptor(s.logger),
		),
		grpc.ConnectionTimeout(s.config.ServerConfig.ShutdownTimeout),
	)

	s.grpcServer = grpc.NewServer(opts...)
	cachev1alpha.RegisterCacheServiceServer(s.grpcServer, s)
	s.metrics.InitializeMetrics(s.grpcServer)
	reflection.Register(s.grpcServer)

	listener, err := s.config.CreateListener()
	if err != nil {
		return err
	}
	s.listener = &listener

	return nil
}

func (s *Server) initHTTPServer() error {
	s.httpServer = &http.Server{
		Addr:              s.config.HTTPListenAddr(),
		Handler:           s.metricsHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	return nil
}

func (s *Server) Init() error {
	s.metrics = grpcprom.NewServerMetrics()
	if err := s.registry.Register(s.metrics); err != nil {
		return err
	}

	if err := s.initGRPCServer(); err != nil {
		return err
	}

	if err := s.initHTTPServer(); err != nil {
		return err
	}

	s.store = store.NewStore(s.config.GetStoreEngine())

	if s.config.IsMemoryStoreDumpEnabled() {
		if err := s.ReadPersistedMemoryStore(); err != nil {
			return err
		}
	}

	return nil
}

func (s *Server) startGRPCServer() error {
	addr := (*s.listener).Addr().String()
	s.logger.Info("Starting gRPC server", "addr", addr)
	return s.grpcServer.Serve(*s.listener)
}

func (s *Server) startHTTPServer() error {
	addr := s.httpServer.Addr
	s.logger.Info("Starting HTTP server", "addr", addr)

	if s.config.TLSConfig != nil && s.config.TLSConfig.Enabled {
		return s.httpServer.ListenAndServeTLS(
			s.config.TLSConfig.CertFile,
			s.config.TLSConfig.KeyFile,
		)
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Start(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- s.startHTTPServer()
	}()

	go func() {
		errCh <- s.startGRPCServer()
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
