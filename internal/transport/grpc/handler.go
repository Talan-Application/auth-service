package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Talan-Application/auth-service/internal/domain"
	authv1 "github.com/Talan-Application/proto-generation/auth/v1"
	"github.com/Talan-Application/auth-service/internal/service"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	authSvc service.AuthService
	log     *zap.Logger
}

func NewHandler(authSvc service.AuthService, log *zap.Logger) *Handler {
	return &Handler{authSvc: authSvc, log: log}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.AuthResponse, error) {
	tokens, err := h.authSvc.Register(ctx,
		req.GetEmail(),
		req.GetPassword(),
		req.GetFirstName(),
		req.GetLastName(),
		req.MiddleName,
	)
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.AuthResponse, error) {
	tokens, err := h.authSvc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.AuthResponse, error) {
	tokens, err := h.authSvc.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *Handler) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	claims, err := h.authSvc.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.ValidateTokenResponse{
		UserId: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (h *Handler) toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInvalidPassword):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrInvalidToken):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrTokenExpired):
		return status.Error(codes.Unauthenticated, err.Error())
	default:
		h.log.Error("unexpected error", zap.Error(err))
		return status.Error(codes.Internal, "internal error")
	}
}
