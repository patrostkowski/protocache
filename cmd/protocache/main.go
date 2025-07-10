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
