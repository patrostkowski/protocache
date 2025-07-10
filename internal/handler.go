package internal

import (
	"context"

	pb "github.com/patrostkowski/protocache/api/pb"
)

type Server struct {
	pb.UnimplementedCacheServiceServer // This is required for forward compatibility
}

func (s *Server) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	// Implement actual logic
	return &pb.SetResponse{Success: true, Message: "Set OK"}, nil
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	return &pb.GetResponse{Found: true, Message: "Value found", Value: []byte("example")}, nil
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	return &pb.DeleteResponse{Success: true, Message: "Deleted"}, nil
}

func (s *Server) Clear(ctx context.Context, req *pb.ClearRequest) (*pb.ClearResponse, error) {
	return &pb.ClearResponse{Success: true, Message: "Cache cleared"}, nil
}
