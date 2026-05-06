package grpc

import (
	"context"
	"errors"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Talan-Application/auth-service/internal/domain"
	"github.com/Talan-Application/auth-service/internal/service"
	authv1 "github.com/Talan-Application/proto-generation/auth/v1"
)

type Handler struct {
	authv1.UnimplementedAuthServiceServer
	authSvc service.AuthService
	log     *zap.Logger
}

func NewHandler(authSvc service.AuthService, log *zap.Logger) *Handler {
	return &Handler{authSvc: authSvc, log: log}
}

func (h *Handler) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.OTPSentResponse, error) {
	err := h.authSvc.Register(ctx,
		req.GetEmail(),
		req.GetPassword(),
		req.GetFirstName(),
		req.GetLastName(),
		req.MiddleName,
	)
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.OTPSentResponse{Message: "verification code sent to your email"}, nil
}

func (h *Handler) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.OTPSentResponse, error) {
	err := h.authSvc.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.OTPSentResponse{Message: "2FA code sent to your email"}, nil
}

func (h *Handler) VerifyEmail(ctx context.Context, req *authv1.VerifyCodeRequest) (*authv1.AuthResponse, error) {
	tokens, err := h.authSvc.VerifyEmail(ctx, req.GetEmail(), req.GetCode())
	if err != nil {
		return nil, h.toGRPCError(err)
	}

	return &authv1.AuthResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *Handler) VerifyLoginCode(ctx context.Context, req *authv1.VerifyCodeRequest) (*authv1.AuthResponse, error) {
	tokens, err := h.authSvc.VerifyLoginCode(ctx, req.GetEmail(), req.GetCode())
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
	case errors.Is(err, domain.ErrInvalidCode):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserNotVerified):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		h.log.Error("unexpected error", zap.Error(err))
		return status.Error(codes.Internal, "internal error")
	}
}
