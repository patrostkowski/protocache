package server

import (
	"context"
	"testing"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/stretchr/testify/assert"
)

func TestSetAndGet(t *testing.T) {
	server := NewTestServer(t)
	ctx := context.Background()

	_, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "foo", Value: []byte("bar")})
	assert.NoError(t, err)

	res, err := server.Get(ctx, &cachev1alpha.GetRequest{Key: "foo"})
	assert.NoError(t, err)
	assert.True(t, res.Found)
	assert.Equal(t, []byte("bar"), res.Value)

	_, err = server.Get(ctx, &cachev1alpha.GetRequest{Key: "baz"})
	assert.Error(t, err)
}

func TestDelete(t *testing.T) {
	server := NewTestServer(t)
	ctx := context.Background()

	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "foo", Value: []byte("bar")}); err != nil {
		t.Fatal(err)
	}

	_, err := server.Delete(ctx, &cachev1alpha.DeleteRequest{Key: "foo"})
	assert.NoError(t, err)

	_, err = server.Get(ctx, &cachev1alpha.GetRequest{Key: "foo"})
	assert.Error(t, err)
}

func TestClear(t *testing.T) {
	server := NewTestServer(t)
	ctx := context.Background()

	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "a", Value: []byte("1")}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "b", Value: []byte("2")}); err != nil {
		t.Fatal(err)
	}

	_, err := server.Clear(ctx, &cachev1alpha.ClearRequest{})
	assert.NoError(t, err)

	_, err = server.Get(ctx, &cachev1alpha.GetRequest{Key: "a"})
	assert.Error(t, err)
	_, err = server.Get(ctx, &cachev1alpha.GetRequest{Key: "b"})
	assert.Error(t, err)
}

func TestList(t *testing.T) {
	server := NewTestServer(t)
	ctx := context.Background()

	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "a", Value: []byte("1")}); err != nil {
		t.Fatal(err)
	}
	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "b", Value: []byte("2")}); err != nil {
		t.Fatal(err)
	}

	resp, err := server.List(ctx, &cachev1alpha.ListRequest{})
	assert.NoError(t, err)
	assert.Contains(t, resp.Keys, "a")
	assert.Contains(t, resp.Keys, "b")

	if _, err := server.Set(ctx, &cachev1alpha.SetRequest{Key: "c", Value: []byte("3")}); err != nil {
		t.Fatal(err)
	}
	resp, err = server.List(ctx, &cachev1alpha.ListRequest{})
	assert.NoError(t, err)
	assert.Contains(t, resp.Keys, "c")

	_, err = server.Clear(ctx, &cachev1alpha.ClearRequest{})
	assert.NoError(t, err)
	resp, err = server.List(ctx, &cachev1alpha.ListRequest{})
	assert.NoError(t, err)
	assert.Empty(t, resp.Keys)
}
