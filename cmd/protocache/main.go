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

package main

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
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/server"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Warn("Could not load config from file, using default config", "error", err)
		cfg = config.DefaultConfig()
	}

	srvMetrics := grpcprom.NewServerMetrics()
	prometheus.MustRegister(srvMetrics)

	httpSrv := &http.Server{Addr: cfg.HTTPListenAddr()}
	m := http.NewServeMux()
	m.Handle("/metrics", promhttp.Handler())
	httpSrv.Handler = m
	go func() {
		logger.Info("starting HTTP server", "addr", httpSrv.Addr)
		if err := httpSrv.ListenAndServe(); err != nil {
			logger.Error("web server error", "err", err)
		}
	}()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			srvMetrics.UnaryServerInterceptor(),
			server.LoggingUnaryInterceptor(logger),
		),
		grpc.ConnectionTimeout(cfg.ServerConfig.ShutdownTimeout),
	)

	cacheService := server.NewServer(logger, cfg)
	if cfg.IsMemoryStoreDumpEnabled() {
		if err := cacheService.ReadPersistedMemoryStore(); err != nil {
			logger.Error("Failed to read the memory store dump", "error", err.Error())
			os.Exit(1)
		}
	}

	pb.RegisterCacheServiceServer(grpcServer, cacheService)
	srvMetrics.InitializeMetrics(grpcServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", cfg.GRPCListenAddr())
	if err != nil {
		logger.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}
	defer lis.Close()
	go func() {
		logger.Info("gRPC server listening", slog.Int("port", cfg.ServerConfig.GRPCPort))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("server error", slog.Any("error", err))
			lis.Close()
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("Signal received, attempting graceful shutdown")
	timer := time.AfterFunc(cfg.ServerConfig.GracefulTimeout, func() {
		logger.Warn("Graceful shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	})
	defer timer.Stop()
	grpcServer.GracefulStop()
	if cfg.IsMemoryStoreDumpEnabled() {
		if err := cacheService.PersistMemoryStore(); err != nil {
			logger.Error("Failed to persist memory store", "error", err)
		}
	}
	logger.Info("Server stopped")
}
