package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"

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

type userRegisteredPayload struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Code   string `json:"code"`
}

type authService struct {
	userRepo  repository.UserRepository
	jwt       *jwt.Manager
	publisher EventPublisher
	log       *zap.Logger
}

func NewAuthService(userRepo repository.UserRepository, cfg config.JWTConfig, publisher EventPublisher, log *zap.Logger) AuthService {
	return &authService{
		userRepo:  userRepo,
		jwt:       jwt.NewManager(cfg.SecretKey, cfg.AccessTokenTTL, cfg.RefreshTokenTTL),
		publisher: publisher,
		log:       log,
	}
}

func (s *authService) Register(ctx context.Context, email, rawPassword, firstName, lastName string, middleName *string) (*TokenPair, error) {
	hash, err := password.Hash(rawPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
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
		return nil, err
	}

	code := generateCode()
	if err := s.publisher.Publish(ctx, "user.registered", userRegisteredPayload{
		UserID: user.ID,
		Email:  user.Email,
		Code:   code,
	}); err != nil {
		s.log.Warn("failed to publish user.registered event", zap.Error(err))
	}

	return s.generateTokenPair(user)
}

func (s *authService) Login(ctx context.Context, email, rawPassword string) (*TokenPair, error) {
	user, err := s.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := password.Verify(user.PasswordHash, rawPassword); err != nil {
		return nil, domain.ErrInvalidPassword
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

// generateCode returns a random 6-digit numeric verification code.
func generateCode() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	n := binary.BigEndian.Uint32(b[:]) % 1_000_000
	return fmt.Sprintf("%06d", n)
}
