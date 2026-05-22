package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/abzalserikbay/jobify/pkg/hasher"
	jwtpkg "github.com/abzalserikbay/jobify/pkg/jwt"
	"github.com/google/uuid"
)

type AuthService struct {
	userRepo      repository.UserRepository
	tokenRepo     repository.TokenRepository
	hasher        *hasher.BcryptHasher
	jwt           *jwtpkg.Manager
	refreshExpiry time.Duration
}

func NewAuthService(
	userRepo repository.UserRepository,
	tokenRepo repository.TokenRepository,
	hasher *hasher.BcryptHasher,
	jwt *jwtpkg.Manager,
	refreshExpiry time.Duration,
) *AuthService {
	return &AuthService{
		userRepo:      userRepo,
		tokenRepo:     tokenRepo,
		hasher:        hasher,
		jwt:           jwt,
		refreshExpiry: refreshExpiry,
	}
}

type RegisterInput struct {
	Email    string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Register(ctx context.Context, in RegisterInput) (*domain.User, error) {
	_, err := s.userRepo.GetByEmail(ctx, in.Email)
	if err == nil {
		return nil, domain.ErrConflict
	}
	if !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	hashed, err := s.hasher.Hash(in.Password)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:       uuid.New(),
		Email:    in.Email,
		Password: hashed,
		Role:     domain.RoleDeveloper,
	}
	profile := &domain.DeveloperProfile{
		ID:     uuid.New(),
		UserID: user.ID,
	}
	if err := s.userRepo.CreateWithProfile(ctx, user, profile); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *AuthService) Login(ctx context.Context, in LoginInput) (*TokenPair, error) {
	user, err := s.userRepo.GetByEmail(ctx, in.Email)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	if err := s.hasher.Compare(in.Password, user.Password); err != nil {
		return nil, domain.ErrUnauthorized
	}

	return s.newTokenPair(ctx, user)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	userID, err := s.tokenRepo.GetUserIDByToken(ctx, refreshToken)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.ErrUnauthorized
	}
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	_ = s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
	return s.newTokenPair(ctx, user)
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
}

func (s *AuthService) newTokenPair(ctx context.Context, user *domain.User) (*TokenPair, error) {
	accessToken, err := s.jwt.Generate(user.ID, string(user.Role))
	if err != nil {
		return nil, err
	}

	refreshToken, err := secureToken()
	if err != nil {
		return nil, err
	}

	if err := s.tokenRepo.StoreRefreshToken(ctx, user.ID, refreshToken, s.refreshExpiry); err != nil {
		return nil, err
	}

	return &TokenPair{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func secureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
