package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/patrostkowski/protocache/api/pb"
	internal "github.com/patrostkowski/protocache/internal"
)

const (
	PORT = 8080
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", PORT))
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(internal.LoggingUnaryInterceptor(logger)),
	)
	cacheService := internal.NewServer()

	pb.RegisterCacheServiceServer(grpcServer, cacheService)
	reflection.Register(grpcServer)

	logger.Info("gRPC server listening", slog.Int("port", PORT))
	if err := grpcServer.Serve(lis); err != nil {
		panic(err)
	}
}
