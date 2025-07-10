package internal

import (
	"context"
	"sync"

	pb "github.com/patrostkowski/protocache/api/pb"
)

type Server struct {
	pb.UnimplementedCacheServiceServer

	mu    sync.RWMutex
	store map[string][]byte
}

func NewServer() *Server {
	return &Server{
		store: make(map[string][]byte),
	}
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
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
		return &pb.GetResponse{Found: false, Message: "not found"}, nil
	}

	return &pb.GetResponse{Found: true, Message: "found", Value: val}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.store, req.Key)
	return &pb.DeleteResponse{Success: true, Message: "deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string][]byte)
	return &pb.ClearResponse{Success: true, Message: "cleared"}, nil
}
