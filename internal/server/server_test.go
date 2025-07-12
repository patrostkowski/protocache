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

package server_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/server"
	testhelpers "github.com/patrostkowski/protocache/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestPersistAndReadMemoryStore(t *testing.T) {
	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "store.gob.gz")
	cfg := &config.Config{MemoryDumpFilePath: dumpPath}

	logger := testhelpers.DefaultLogger()

	s1 := server.NewServer(logger, cfg)
	ctx := context.Background()

	_, err := s1.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})
	require.NoError(t, err)

	_, err = s1.Set(ctx, &pb.SetRequest{Key: "baz", Value: []byte("qux")})
	require.NoError(t, err)

	err = s1.PersistMemoryStore()
	require.NoError(t, err)

	s2 := server.NewServer(logger, cfg)
	err = s2.ReadPersistedMemoryStore()
	require.NoError(t, err)

	resp, err := s2.Get(ctx, &pb.GetRequest{Key: "foo"})
	require.NoError(t, err)
	assert.Equal(t, []byte("bar"), resp.Value)

	resp, err = s2.Get(ctx, &pb.GetRequest{Key: "baz"})
	require.NoError(t, err)
	assert.Equal(t, []byte("qux"), resp.Value)
}

func TestReadPersistedMemoryStore_FileNotFound(t *testing.T) {
	s := testhelpers.NewTestServer(t)

	err := s.ReadPersistedMemoryStore()
	assert.NoError(t, err)
}

func TestReadPersistedMemoryStore_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "store.gob.gz")

	err := os.WriteFile(dumpPath, []byte{}, 0600)
	require.NoError(t, err)

	cfg := &config.Config{
		MemoryDumpFilePath: dumpPath,
	}

	s := server.NewServer(testhelpers.DefaultLogger(), cfg)
	err = s.ReadPersistedMemoryStore()
	assert.NoError(t, err)
}
