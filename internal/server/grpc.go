package server

import (
	"context"
	"errors"
	"runtime"
	"time"

	pb "github.com/patrostkowski/protocache/api/pb"
	"github.com/patrostkowski/protocache/internal/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key must not be empty")
	}

	if err := s.store.Set(req.Key, req.Value); err != nil {
		return nil, status.Errorf(codes.Aborted, "could not set %q key", req.Key)
	}
	return &pb.SetResponse{Success: true, Message: "OK"}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	val, err := s.store.Get(req.Key)
	if err != nil {
		if errors.Is(err, store.StoreErrorKeyNotFound) {
			CacheMisses.Inc()
			return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
		}
		return nil, status.Errorf(codes.Unknown, "internal error: %v", err)
	}

	CacheHits.Inc()
	return &pb.GetResponse{Found: true, Message: "found", Value: val}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	if err := s.store.Delete(req.Key); err != nil {
		return nil, status.Errorf(codes.Unknown, "internal error: %v", err)
	}
	return &pb.DeleteResponse{Success: true, Message: "deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.store.Clear()
	return &pb.ClearResponse{Success: true, Message: "cleared"}, nil
}

func (s *Server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	keys := s.store.List()
	return &pb.ListResponse{Keys: keys}, nil
}

func (s *Server) Stats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsResponse, error) {
	var totalBytes uint64
	thisStore := s.store.This()
	for _, v := range thisStore {
		totalBytes += uint64(len(v))
	}

	return &pb.StatsResponse{
		KeyCount:         uint64(len(thisStore)),
		MemoryUsageBytes: totalBytes,
		GoVersion:        runtime.Version(),
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}
