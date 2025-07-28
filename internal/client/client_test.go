package client

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// mockServer implements the CacheServiceServer interface for test purposes.
type mockServer struct {
	v1alpha.UnimplementedCacheServiceServer
	store map[string][]byte
}

func newMockServer() *mockServer {
	return &mockServer{store: make(map[string][]byte)}
}

func (s *mockServer) Set(ctx context.Context, req *v1alpha.SetRequest) (*v1alpha.SetResponse, error) {
	s.store[req.Key] = req.Value
	return &v1alpha.SetResponse{}, nil
}

func (s *mockServer) Get(ctx context.Context, req *v1alpha.GetRequest) (*v1alpha.GetResponse, error) {
	val, ok := s.store[req.Key]
	return &v1alpha.GetResponse{Value: val, Found: ok}, nil
}

func (s *mockServer) Delete(ctx context.Context, req *v1alpha.DeleteRequest) (*v1alpha.DeleteResponse, error) {
	delete(s.store, req.Key)
	return &v1alpha.DeleteResponse{}, nil
}

func (s *mockServer) Clear(ctx context.Context, req *v1alpha.ClearRequest) (*v1alpha.ClearResponse, error) {
	s.store = make(map[string][]byte)
	return &v1alpha.ClearResponse{}, nil
}

func (s *mockServer) List(ctx context.Context, req *v1alpha.ListRequest) (*v1alpha.ListResponse, error) {
	var keys []string
	for k := range s.store {
		keys = append(keys, k)
	}
	return &v1alpha.ListResponse{Keys: keys}, nil
}

func (s *mockServer) Stats(ctx context.Context, req *v1alpha.StatsRequest) (*v1alpha.StatsResponse, error) {
	return &v1alpha.StatsResponse{KeyCount: uint64(len(s.store))}, nil
}

func startMockServer(t *testing.T) (addr string, stop func()) {
	t.Helper()
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	v1alpha.RegisterCacheServiceServer(grpcServer, newMockServer())

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	return lis.Addr().String(), func() {
		grpcServer.Stop()
		lis.Close()
	}
}

func TestClient(t *testing.T) {
	addr, stop := startMockServer(t)
	defer stop()

	cfg := Config{
		Host:    "127.0.0.1",
		Port:    parsePort(addr),
		Timeout: 2 * time.Second,
	}

	c, err := New(cfg)
	require.NoError(t, err)
	defer c.Close()

	ctx := context.Background()

	// Test Set and Get
	err = c.Set(ctx, "foo", "bar")
	require.NoError(t, err)

	res, err := c.Get(ctx, "foo")
	require.NoError(t, err)
	require.True(t, res.Found)
	require.Equal(t, "bar", string(res.Value))

	// Test Delete
	err = c.Delete(ctx, "foo")
	require.NoError(t, err)

	res, err = c.Get(ctx, "foo")
	require.NoError(t, err)
	require.False(t, res.Found)

	// Test Set multiple + List
	_ = c.Set(ctx, "a", "1")
	_ = c.Set(ctx, "b", "2")
	keys, err := c.List(ctx)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"a", "b"}, keys)

	// Test Stats
	stats, err := c.Stats(ctx)
	require.NoError(t, err)
	require.Equal(t, uint64(2), stats.KeyCount)

	// Test Clear
	err = c.Clear(ctx)
	require.NoError(t, err)
	keys, err = c.List(ctx)
	require.NoError(t, err)
	require.Empty(t, keys)
}

func parsePort(addr string) int {
	_, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	var port int
	_, _ = fmt.Sscanf(portStr, "%d", &port)
	return port
}
