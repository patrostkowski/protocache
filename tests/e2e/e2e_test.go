package internal_test

import (
	"context"
	"net"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func startTestServer(t *testing.T) (addr string, stop func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:8080")
	assert.NoError(t, err)

	server := grpc.NewServer()
	cacheService := internal.NewServer()
	pb.RegisterCacheServiceServer(server, cacheService)

	go func() {
		_ = server.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		server.Stop()
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

	client := pb.NewCacheServiceClient(conn)
	ctx := context.Background()

	_, err = client.Set(ctx, &pb.SetRequest{Key: "test", Value: []byte("hello")})
	assert.NoError(t, err)

	getResp, err := client.Get(ctx, &pb.GetRequest{Key: "test"})
	assert.NoError(t, err)
	assert.True(t, getResp.Found)
	assert.Equal(t, []byte("hello"), getResp.Value)

	_, err = client.Delete(ctx, &pb.DeleteRequest{Key: "test"})
	assert.NoError(t, err)

	getResp, err = client.Get(ctx, &pb.GetRequest{Key: "test"})
	assert.NoError(t, err)
	assert.False(t, getResp.Found)
}
