package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/Talan-Application/auth-service/internal/config"
	"github.com/Talan-Application/auth-service/internal/domain"
	"github.com/Talan-Application/auth-service/internal/repository"
	"github.com/Talan-Application/auth-service/pkg/jwt"
	"github.com/Talan-Application/auth-service/pkg/password"
)

// EventPublisher publishes domain events to the message broker.
type EventPublisher interface {
	Publish(ctx context.Context, routingKey string, payload any) error
}

type otpPayload struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Code   string `json:"code"`
}

type authService struct {
	userRepo repository.UserRepository
	codeRepo repository.CodeRepository
	jwt      *jwt.Manager
	publisher EventPublisher
	log      *zap.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	codeRepo repository.CodeRepository,
	cfg config.JWTConfig,
	publisher EventPublisher,
	log *zap.Logger,
) AuthService {
	return &authService{
		userRepo:  userRepo,
		codeRepo:  codeRepo,
		jwt:       jwt.NewManager(cfg.SecretKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		publisher: publisher,
		log:       log,
	}
}

func (s *authService) Register(ctx context.Context, email, rawPassword, firstName, lastName string, middleName *string) error {
	hash, err := password.Hash(rawPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		Email:        email,
		FirstName:    firstName,
		LastName:     lastName,
		MiddleName:   middleName,
		PasswordHash: hash,
		Role:         domain.RoleStudent,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	return s.sendOTP(ctx, user.ID, email, domain.CodePurposeEmailVerification, "user.registered")
}

func (s *authService) Login(ctx context.Context, email, rawPassword string) error {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return err
	}

	if err := password.Verify(user.PasswordHash, rawPassword); err != nil {
		return domain.ErrInvalidPassword
	}

	if !user.IsVerified {
		return domain.ErrUserNotVerified
	}

	return s.sendOTP(ctx, user.ID, email, domain.CodePurposeLoginOTP, "user.login_otp")
}

func (s *authService) VerifyEmail(ctx context.Context, email, code string) (*TokenPair, error) {
	c, err := s.codeRepo.FindValid(ctx, email, code, domain.CodePurposeEmailVerification)
	if err != nil {
		return nil, err
	}

	if err := s.codeRepo.MarkUsed(ctx, c.ID); err != nil {
		s.log.Warn("failed to mark code used", zap.Error(err))
	}

	if err := s.userRepo.SetVerified(ctx, email); err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

func (s *authService) VerifyLoginCode(ctx context.Context, email, code string) (*TokenPair, error) {
	c, err := s.codeRepo.FindValid(ctx, email, code, domain.CodePurposeLoginOTP)
	if err != nil {
		return nil, err
	}

	if err := s.codeRepo.MarkUsed(ctx, c.ID); err != nil {
		s.log.Warn("failed to mark code used", zap.Error(err))
	}

	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error) {
	claims, err := s.jwt.Validate(refreshToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(user)
}

func (s *authService) ValidateToken(_ context.Context, accessToken string) (*Claims, error) {
	claims, err := s.jwt.Validate(accessToken)
	if err != nil {
		return nil, domain.ErrInvalidToken
	}

	return &Claims{
		UserID: claims.UserID,
		Email:  claims.Email,
		Role:   claims.Role,
	}, nil
}

func (s *authService) sendOTP(ctx context.Context, userID int64, email, purpose, routingKey string) error {
	code := generateCode()

	if err := s.codeRepo.Create(ctx, &domain.Code{
		Code:      code,
		Receiver:  email,
		Purpose:   purpose,
		ExpiredAt: time.Now().Add(10 * time.Minute),
	}); err != nil {
		return fmt.Errorf("save otp code: %w", err)
	}

	if err := s.publisher.Publish(ctx, routingKey, otpPayload{
		UserID: userID,
		Email:  email,
		Code:   code,
	}); err != nil {
		s.log.Warn("failed to publish otp event", zap.String("routing_key", routingKey), zap.Error(err))
	}

	return nil
}

func (s *authService) generateTokenPair(user *domain.User) (*TokenPair, error) {
	accessToken, err := s.jwt.GenerateAccessToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwt.GenerateRefreshToken(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func generateCode() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	n := binary.BigEndian.Uint32(b[:]) % 1_000_000
	return fmt.Sprintf("%06d", n)
}
