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
	"flag"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	grpcprom "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/logger"
	"github.com/patrostkowski/protocache/internal/store"
	cachev1alpha "github.com/patrostkowski/protocache/pkg/api/cache/v1alpha"
)

type Server struct {
	cachev1alpha.UnimplementedCacheServiceServer

	store      store.Store
	config     *config.Config
	listener   *net.Listener
	grpcServer *grpc.Server
	httpServer *http.Server
	metrics    *grpcprom.ServerMetrics
	registry   prometheus.Registerer
}

func NewServer(config *config.Config, reg prometheus.Registerer) *Server {
	if reg == nil {
		reg = prometheus.DefaultRegisterer
	}
	store := store.NewStore(config.GetStoreEngine(), config.GetEvictionPolicy())
	return &Server{
		store:    store,
		config:   config,
		registry: reg,
		metrics:  grpcprom.NewServerMetrics(),
	}
}

func (s *Server) initGRPCServer() error {
	opts := []grpc.ServerOption{}

	if tlsOpt, err := s.config.CreateTLS(); err != nil {
		return err
	} else if tlsOpt != nil {
		logger.Info("TLS credentials initialized",
			"cert", s.config.TLSConfig.CertFile,
			"key", s.config.TLSConfig.KeyFile,
		)

		opts = append(opts, tlsOpt)
	}

	if opt := s.config.GRPCConnectionTimeoutOption(); opt != nil {
		opts = append(opts, opt)
	}

	opts = append(opts,
		grpc.ChainUnaryInterceptor(
			s.metrics.UnaryServerInterceptor(),
			LoggingUnaryInterceptor(),
		),
	)

	s.grpcServer = grpc.NewServer(opts...)
	cachev1alpha.RegisterCacheServiceServer(s.grpcServer, s)
	s.metrics.InitializeMetrics(s.grpcServer)
	reflection.Register(s.grpcServer)

	listener, err := s.config.CreateListener()
	if err != nil {
		logger.Error("Failed to create listener", "error", err)
		return err
	}
	s.listener = &listener
	logger.Info("Listener created", "address", listener.Addr().String())

	return nil
}

func (s *Server) initHTTPServer() error {
	s.httpServer = &http.Server{
		Addr:              s.config.HTTPListenAddr(),
		Handler:           s.metricsHandler(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	logger.Info("HTTP server initialized", "addr", s.config.HTTPListenAddr())
	return nil
}

func (s *Server) Init() error {
	s.metrics = grpcprom.NewServerMetrics()
	if err := s.registry.Register(s.metrics); err != nil {
		return err
	}

	if err := s.initGRPCServer(); err != nil {
		logger.Error("gRPC server initialization failed", "error", err)
		return err
	}
	logger.Debug("gRPC server initialized successfully")

	if err := s.initHTTPServer(); err != nil {
		logger.Error("HTTP server initialization failed", "error", err)
		return err
	}
	logger.Debug("HTTP server initialized successfully")

	if s.config.IsMemoryStoreDumpEnabled() {
		logger.Info("Memory store dump is enabled. Attempting to restore from disk")
		if err := s.ReadPersistedMemoryStore(); err != nil {
			logger.Error("Failed to read persisted memory store", "error", err)
			return err
		}
		logger.Info("Successfully read memory store dump into memory")
	}

	return nil
}

func (s *Server) startGRPCServer() error {
	addr := (*s.listener).Addr().String()
	logger.Info("Starting gRPC server", "addr", addr)
	return s.grpcServer.Serve(*s.listener)
}

func (s *Server) startHTTPServer() error {
	addr := s.httpServer.Addr
	logger.Info("Starting HTTP server", "addr", addr)

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
		logger.Info("Shutdown signal received")
	case err := <-errCh:
		logger.Error("Server crashed", "error", err)
		return err
	}

	return s.Shutdown()
}

func (s *Server) Shutdown() error {
	logger.Info("Initiating shutdown sequence")

	shutdownTimer := time.AfterFunc(s.config.ServerConfig.GracefulTimeout, func() {
		logger.Warn("Graceful shutdown timeout exceeded, forcing stop")
		s.grpcServer.Stop()
	})
	defer shutdownTimer.Stop()

	s.grpcServer.GracefulStop()

	if err := s.httpServer.Shutdown(context.Background()); err != nil {
		logger.Error("Error shutting down HTTP server", "error", err)
	}

	if s.config.IsMemoryStoreDumpEnabled() {
		if err := s.PersistMemoryStore(); err != nil {
			logger.Error("Failed to persist memory store", "error", err)
			return err
		}
	}

	logger.Info("Server shut down cleanly")
	return nil
}

func Run() error {
	var configPath string
	flag.StringVar(&configPath, "config", config.ConfigFileDefaultFilePath, "Path to configuration file")
	flag.Parse()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		logger.Warn("using default config", "error", err)
		cfg = config.DefaultConfig()
	}

	srv := NewServer(cfg, prometheus.DefaultRegisterer)

	if err := srv.Init(); err != nil {
		logger.Error("Initialization failed", "error", err)
		return err
	}

	if err := srv.Start(ctx); err != nil {
		return err
	}

	return nil
}
