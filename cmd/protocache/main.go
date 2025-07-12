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
	"fmt"
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
	internal "github.com/patrostkowski/protocache/internal"
)

const (
	GRPC_PORT               = 8081
	METRICS_PORT            = 8080
	SERVER_SHUTDOWN_TIMEOUT = 30 * time.Second
	GRACEFUL_TIMEOUT_SEC    = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	srvMetrics := grpcprom.NewServerMetrics()
	prometheus.MustRegister(srvMetrics)

	httpSrv := &http.Server{Addr: "0.0.0.0:8080"}
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
			internal.LoggingUnaryInterceptor(logger),
		),
		grpc.ConnectionTimeout(SERVER_SHUTDOWN_TIMEOUT),
	)
	cacheService := internal.NewServer()
	pb.RegisterCacheServiceServer(grpcServer, cacheService)
	srvMetrics.InitializeMetrics(grpcServer)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", GRPC_PORT))
	if err != nil {
		logger.Error("failed to listen", slog.Any("error", err))
		os.Exit(1)
	}
	go func() {
		logger.Info("gRPC server listening", slog.Int("port", GRPC_PORT))
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("server error", slog.Any("error", err))
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	stop()
	logger.Info("Signal received, attempting graceful shutdown")
	timer := time.AfterFunc(GRACEFUL_TIMEOUT_SEC, func() {
		logger.Warn("Graceful shutdown timeout exceeded, forcing stop")
		grpcServer.Stop()
	})
	defer timer.Stop()
	grpcServer.GracefulStop()
	logger.Info("Server stopped")
}
