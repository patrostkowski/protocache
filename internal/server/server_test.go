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
	"time"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/config"
	"github.com/patrostkowski/protocache/internal/server"
	testhelpers "github.com/patrostkowski/protocache/internal/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersistAndReadMemoryStore(t *testing.T) {
	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "store.gob.gz")
	cfg := &config.Config{
		StoreConfig: config.StoreConfig{
			MemoryDumpPath: dumpPath,
		},
	}

	logger := testhelpers.DefaultLogger()

	s1 := server.NewServer(logger, cfg, testhelpers.DefaultPrometheusRegistry())
	ctx := context.Background()

	_, err := s1.Set(ctx, &pb.SetRequest{Key: "foo", Value: []byte("bar")})
	require.NoError(t, err)

	_, err = s1.Set(ctx, &pb.SetRequest{Key: "baz", Value: []byte("qux")})
	require.NoError(t, err)

	err = s1.PersistMemoryStore()
	require.NoError(t, err)

	s2 := server.NewServer(logger, cfg, testhelpers.DefaultPrometheusRegistry())
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
		StoreConfig: config.StoreConfig{
			MemoryDumpPath: dumpPath,
		},
	}

	s := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	err = s.ReadPersistedMemoryStore()
	assert.NoError(t, err)
}

func TestServerLifecycle_InitAndShutdown(t *testing.T) {
	cfg := &config.Config{
		ServerConfig: config.ServerConfig{
			ShutdownTimeout: 1 * time.Second,
			HTTPPort:        0,
			GRPCPort:        0,
		},
	}

	s := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())

	err := s.Init()
	assert.NoError(t, err)

	err = s.Shutdown()
	assert.NoError(t, err)
}

func TestServerStart_CancelContextTriggersShutdown(t *testing.T) {
	cfg := &config.Config{
		ServerConfig: config.ServerConfig{
			ShutdownTimeout: 1 * time.Second,
			HTTPPort:        0,
			GRPCPort:        0,
		},
	}

	s := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	require.NoError(t, s.Init())

	t.Cleanup(func() {
		err := s.Shutdown()
		require.NoError(t, err)
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)

	go func() {
		done <- s.Start(ctx)
	}()

	ready := make(chan struct{})
	go func() {
		time.Sleep(100 * time.Millisecond)
		close(ready)
	}()
	<-ready

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Server did not shut down in time")
	}
}

func TestServerInit_ReadPersistedMemoryStoreFails(t *testing.T) {
	tmpDir := t.TempDir()
	dumpPath := filepath.Join(tmpDir, "bad-store.gob.gz")

	// Write corrupted file
	require.NoError(t, os.WriteFile(dumpPath, []byte("not a valid gob"), 0600))

	cfg := &config.Config{
		StoreConfig: config.StoreConfig{
			DumpEnabled:    true,
			MemoryDumpPath: dumpPath,
		},
		ServerConfig: config.ServerConfig{
			ShutdownTimeout: 1 * time.Second,
		},
	}

	s := server.NewServer(testhelpers.DefaultLogger(), cfg, testhelpers.DefaultPrometheusRegistry())
	err := s.Init()
	assert.Error(t, err)
}
