package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb "github.com/patrostkowski/protocache/api/pb"
	internal "github.com/patrostkowski/protocache/internal"
)

const (
	PORT = 8080
)

func main() {
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", PORT))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	cacheService := &internal.Server{}

	pb.RegisterCacheServiceServer(grpcServer, cacheService)
	reflection.Register(grpcServer)

	log.Printf("gRPC server listening on %d\n", PORT)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
