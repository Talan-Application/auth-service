package grpc

import (
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Talan-Application/auth-service/internal/config"
	"github.com/Talan-Application/auth-service/internal/service"
	authv1 "github.com/Talan-Application/proto-generation/auth/v1"
)

type Server struct {
	grpcServer *grpc.Server
	port       int
	log        *zap.Logger
}

func NewServer(cfg config.GRPCConfig, log *zap.Logger, authSvc service.AuthService) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
			recoveryInterceptor(log),
		),
	)

	handler := NewHandler(authSvc, log)
	authv1.RegisterAuthServiceServer(grpcServer, handler)

	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		port:       cfg.Port,
		log:        log,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	s.log.Info("gRPC server started", zap.Int("port", s.port))
	return s.grpcServer.Serve(lis)
}

func (s *Server) GracefulStop() {
	s.grpcServer.GracefulStop()
	s.log.Info("gRPC server stopped")
}
