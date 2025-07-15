package server_test

import (
	"context"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	testhelpers "github.com/patrostkowski/protocache/internal/test"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGet(t *testing.T) {
	server := testhelpers.NewTestServer(t)
	ctx := context.Background()

	_, err := server.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})
	assert.NoError(t, err)

	res, err := server.Get(ctx, &pb.GetRequest{Key: "foo"})
	assert.NoError(t, err)
	assert.True(t, res.Found)
	assert.Equal(t, []byte("bar"), res.Value)

	_, err = server.Get(ctx, &pb.GetRequest{Key: "baz"})
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	server := testhelpers.NewTestServer(t)
	ctx := context.Background()

	server.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})

	_, err := server.Delete(ctx, &pb.DeleteRequest{Key: "foo"})
	assert.NoError(t, err)

	_, err = server.Get(ctx, &pb.GetRequest{Key: "foo"})
	assert.Error(t, err)
}

func TestClear(t *testing.T) {
	server := testhelpers.NewTestServer(t)
	ctx := context.Background()

	server.Set(ctx, &pb.SetRequest{Key: "a", Value: []byte("1")})
	server.Set(ctx, &pb.SetRequest{Key: "b", Value: []byte("2")})

	_, err := server.Clear(ctx, &pb.ClearRequest{})
	assert.NoError(t, err)

	_, err = server.Get(ctx, &pb.GetRequest{Key: "a"})
	assert.Error(t, err)
	_, err = server.Get(ctx, &pb.GetRequest{Key: "b"})
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	server := testhelpers.NewTestServer(t)
	ctx := context.Background()

	server.Set(ctx, &pb.SetRequest{Key: "a", Value: []byte("1")})
	server.Set(ctx, &pb.SetRequest{Key: "b", Value: []byte("2")})

	resp, err := server.List(ctx, &pb.ListRequest{})
	assert.NoError(t, err)
	assert.Contains(t, resp.Keys, "a")
	assert.Contains(t, resp.Keys, "b")

	server.Set(ctx, &pb.SetRequest{Key: "c", Value: []byte("3")})
	resp, err = server.List(ctx, &pb.ListRequest{})
	assert.NoError(t, err)
	assert.Contains(t, resp.Keys, "c")

	_, err = server.Clear(ctx, &pb.ClearRequest{})
	assert.NoError(t, err)
	resp, err = server.List(ctx, &pb.ListRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.Keys)
}
