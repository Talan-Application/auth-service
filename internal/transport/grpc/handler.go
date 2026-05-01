package grpc

import (
	"github.com/Talan-Application/auth-service/internal/service"
)

// Handler implements the proto-generated AuthServiceServer interface.
// Methods are added here after proto generation via GitHub Actions.
type Handler struct {
	authSvc service.AuthService
}

func NewHandler(authSvc service.AuthService) *Handler {
	return &Handler{authSvc: authSvc}
}
