package internal

import (
	"context"
	"log/slog"
	"sync"

	pb "github.com/patrostkowski/protocache/api/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
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

func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		resp, err := handler(ctx, req)

		var remoteAddr string
		if p, ok := peer.FromContext(ctx); ok {
			remoteAddr = p.Addr.String()
		}

		st, _ := status.FromError(err)

		var key string
		switch r := req.(type) {
		case *pb.SetRequest:
			key = r.GetKey()
		case *pb.GetRequest:
			key = r.GetKey()
		case *pb.DeleteRequest:
			key = r.GetKey()
		}

		logger.Info("gRPC request",
			slog.String("method", info.FullMethod),
			slog.String("remote", remoteAddr),
			slog.String("key", key),
			slog.String("code", st.Code().String()),
		)

		return resp, err
	}
}
