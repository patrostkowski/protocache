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

package server

import (
	"context"
	"errors"
	"runtime"
	"time"

	cachev1alpha "github.com/patrostkowski/protocache/internal/api/cache/v1alpha"
	"github.com/patrostkowski/protocache/internal/logger"
	"github.com/patrostkowski/protocache/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Set(ctx context.Context, req *cachev1alpha.SetRequest) (*cachev1alpha.SetResponse, error) {
	if req.Key == "" {
		logger.Error("Failed to set empty key in store")
		return nil, status.Error(codes.InvalidArgument, "key must not be empty")
	}

	if err := s.store.Set(req.Key, req.Value); err != nil {
		logger.Error("Failed to set key in store", "key", req.Key, "error", err)
		return nil, status.Errorf(codes.Aborted, "could not set %q key", req.Key)
	}
	return &cachev1alpha.SetResponse{Success: true, Message: "OK"}, nil
}

func (s *Server) Get(ctx context.Context, req *cachev1alpha.GetRequest) (*cachev1alpha.GetResponse, error) {
	val, err := s.store.Get(req.Key)
	if err != nil {
		if errors.Is(err, store.StoreErrorKeyNotFound) {
			CacheMisses.Inc()
			logger.Warn("Cache miss", "key", req.Key)
			return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
		}
		logger.Error("Failed to get key from store", "key", req.Key, "error", err)
		return nil, status.Errorf(codes.Unknown, "internal error: %v", err)
	}

	CacheHits.Inc()
	return &cachev1alpha.GetResponse{Found: true, Message: "found", Value: val}, nil
}

func (s *Server) Delete(ctx context.Context, req *cachev1alpha.DeleteRequest) (*cachev1alpha.DeleteResponse, error) {
	if err := s.store.Delete(req.Key); err != nil {
		logger.Error("Failed to delete key from store", "key", req.Key, "error", err)
		return nil, status.Errorf(codes.Unknown, "internal error: %v", err)
	}
	return &cachev1alpha.DeleteResponse{Success: true, Message: "deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *cachev1alpha.ClearRequest) (*cachev1alpha.ClearResponse, error) {
	s.store.Clear()
	return &cachev1alpha.ClearResponse{Success: true, Message: "cleared"}, nil
}

func (s *Server) List(ctx context.Context, req *cachev1alpha.ListRequest) (*cachev1alpha.ListResponse, error) {
	keys := s.store.List()
	return &cachev1alpha.ListResponse{Keys: keys}, nil
}

func (s *Server) Stats(ctx context.Context, _ *cachev1alpha.StatsRequest) (*cachev1alpha.StatsResponse, error) {
	var totalBytes uint64
	thisStore := s.store.This()
	for _, v := range thisStore {
		totalBytes += uint64(len(v))
	}

	return &cachev1alpha.StatsResponse{
		KeyCount:         uint64(len(thisStore)),
		MemoryUsageBytes: totalBytes,
		GoVersion:        runtime.Version(),
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}
