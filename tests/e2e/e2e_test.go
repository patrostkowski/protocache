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

package internal_test

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/server"
)

func startTestServer(t *testing.T) (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

	cfg := config.DefaultConfig()
	cacheService := server.NewServer(server.DefaultLogger(), cfg, server.DefaultPrometheusRegistry())
	assert.NoError(t, err)

	grpcServer := grpc.NewServer()
	cachev1alpha.RegisterCacheServiceServer(grpcServer, cacheService)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcServer.Stop()
		lis.Close()
	}
}

func TestCacheService_E2E(t *testing.T) {
	addr, shutdown := startTestServer(t)
	defer shutdown()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()
	assert.NoError(t, err)

	client := cachev1alpha.NewCacheServiceClient(conn)
	ctx := context.Background()

	_, err = client.Set(ctx, &cachev1alpha.SetRequest{Key: "test", Value: []byte("hello")})
	assert.NoError(t, err)

	getResp, err := client.Get(ctx, &cachev1alpha.GetRequest{Key: "test"})
	assert.NoError(t, err)
	assert.True(t, getResp.Found)
	assert.Equal(t, []byte("hello"), getResp.Value)

	_, err = client.Delete(ctx, &cachev1alpha.DeleteRequest{Key: "test"})
	assert.NoError(t, err)

	getResp, err = client.Get(ctx, &cachev1alpha.GetRequest{Key: "test"})
	assert.Error(t, err)
	assert.Nil(t, getResp)

	st, ok := status.FromError(err)
	assert.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
}
