package server

import (
	"context"
	"maps"
	"runtime"
	"slices"
	"time"

	pb "github.com/patrostkowski/protocache/api/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	if req.Key == "" {
		return nil, status.Error(codes.InvalidArgument, "key must not be empty")
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[req.Key] = req.Value
	return &pb.SetResponse{Success: true, Message: "OK"}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[req.Key]

	if !ok {
		CacheMisses.Inc()
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
	}

	CacheHits.Inc()
	return &pb.GetResponse{Found: true, Message: "found", Value: val}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.store[req.Key]; !ok {
		return nil, status.Errorf(codes.NotFound, "key %q not found", req.Key)
	}

	delete(s.store, req.Key)
	return &pb.DeleteResponse{Success: true, Message: "deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string][]byte)
	return &pb.ClearResponse{Success: true, Message: "cleared"}, nil
}

func (s *Server) List(ctx context.Context, req *pb.ListRequest) (*pb.ListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := slices.Collect(maps.Keys(s.store))
	return &pb.ListResponse{Keys: keys}, nil
}

func (s *Server) Stats(ctx context.Context, _ *pb.StatsRequest) (*pb.StatsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var totalBytes uint64
	for _, v := range s.store {
		totalBytes += uint64(len(v))
	}

	return &pb.StatsResponse{
		KeyCount:         uint64(len(s.store)),
		MemoryUsageBytes: totalBytes,
		GoVersion:        runtime.Version(),
		Timestamp:        time.Now().Format(time.RFC3339),
	}, nil
}
