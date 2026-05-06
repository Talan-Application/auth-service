package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/Talan-Application/auth-service/internal/config"
	"github.com/Talan-Application/auth-service/internal/repository/postgres"
	"github.com/Talan-Application/auth-service/internal/service"
	grpcserver "github.com/Talan-Application/auth-service/internal/transport/grpc"
	"github.com/Talan-Application/auth-service/pkg/logger"
	"github.com/Talan-Application/auth-service/pkg/rabbitmq"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	zapLog := logger.New(cfg.App.Env)
	defer zapLog.Sync()

	db, err := postgres.NewConnection(cfg.Database)
	if err != nil {
		zapLog.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQ.URL, zapLog)
	if err != nil {
		zapLog.Fatal("failed to connect to rabbitmq", zap.Error(err))
	}
	defer rmqConn.Close()

	publisher := rabbitmq.NewPublisher(rmqConn)

	userRepo := postgres.NewUserRepository(db)
	codeRepo := postgres.NewCodeRepository(db)
	authSvc := service.NewAuthService(userRepo, codeRepo, cfg.JWT, publisher, zapLog)

	srv := grpcserver.NewServer(cfg.GRPC, zapLog, authSvc)

	go func() {
		if err := srv.Run(); err != nil {
			zapLog.Fatal("grpc server failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	srv.GracefulStop()
	zapLog.Info("server shut down gracefully")
}
