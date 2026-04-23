package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/pkg/hasher"
	jwtpkg "github.com/abzalserikbay/jobify/pkg/jwt"
	"github.com/google/uuid"
)

type mockUserRepo struct {
	getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	createFn     func(ctx context.Context, u *domain.User) error
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.getByEmailFn(ctx, email)
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockUserRepo) Create(ctx context.Context, u *domain.User) error {
	return m.createFn(ctx, u)
}

type mockProfileRepo struct {
	createFn      func(ctx context.Context, p *domain.DeveloperProfile) error
	getByUserIDFn func(ctx context.Context, userID uuid.UUID) (*domain.DeveloperProfile, error)
	updateFn      func(ctx context.Context, p *domain.DeveloperProfile) error
}

func (m *mockProfileRepo) Create(ctx context.Context, p *domain.DeveloperProfile) error {
	return m.createFn(ctx, p)
}

func (m *mockProfileRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.DeveloperProfile, error) {
	return m.getByUserIDFn(ctx, userID)
}

func (m *mockProfileRepo) Update(ctx context.Context, p *domain.DeveloperProfile) error {
	return m.updateFn(ctx, p)
}

func TestAuthService_Register(t *testing.T) {
	errDB := errors.New("db down")

	tests := []struct {
		name         string
		in           RegisterInput
		getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
		createUserFn func(ctx context.Context, u *domain.User) error
		createProfFn func(ctx context.Context, p *domain.DeveloperProfile) error
		wantErr      error
	}{
		{
			name: "success",
			in:   RegisterInput{Email: "new@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			createUserFn: func(ctx context.Context, u *domain.User) error { return nil },
			createProfFn: func(ctx context.Context, p *domain.DeveloperProfile) error { return nil },
			wantErr:      nil,
		},
		{
			name: "email conflict",
			in:   RegisterInput{Email: "dup@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: uuid.New(), Email: email}, nil
			},
			wantErr: domain.ErrConflict,
		},
		{
			name: "repo error on get",
			in:   RegisterInput{Email: "x@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errDB
			},
			wantErr: errDB,
		},
		{
			name: "repo error on create",
			in:   RegisterInput{Email: "y@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			createUserFn: func(ctx context.Context, u *domain.User) error { return errDB },
			wantErr:      errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{getByEmailFn: tt.getByEmailFn, createFn: tt.createUserFn}
			profRepo := &mockProfileRepo{createFn: tt.createProfFn}
			svc := NewAuthService(userRepo, profRepo, hasher.New(), jwtpkg.NewManager("test", time.Hour))

			user, err := svc.Register(context.Background(), tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if user == nil {
					t.Fatal("expected user, got nil")
				}
				if user.Email != tt.in.Email {
					t.Errorf("email: got %q, want %q", user.Email, tt.in.Email)
				}
				if user.Password == tt.in.Password {
					t.Error("password must be hashed, not stored as plaintext")
				}
				if user.Role != domain.RoleDeveloper {
					t.Errorf("role: got %q, want %q", user.Role, domain.RoleDeveloper)
				}
			}
		})
	}
}

func TestAuthService_Login(t *testing.T) {
	h := hasher.New()
	correctHash, err := h.Hash("correct-password")
	if err != nil {
		t.Fatalf("hash setup: %v", err)
	}
	uid := uuid.New()
	errDB := errors.New("db down")

	tests := []struct {
		name         string
		in           LoginInput
		getByEmailFn func(ctx context.Context, email string) (*domain.User, error)
		wantErr      error
		wantToken    bool
	}{
		{
			name: "success",
			in:   LoginInput{Email: "u@example.com", Password: "correct-password"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
			},
			wantErr:   nil,
			wantToken: true,
		},
		{
			name: "wrong password",
			in:   LoginInput{Email: "u@example.com", Password: "nope"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
			},
			wantErr: domain.ErrUnauthorized,
		},
		{
			name: "user not found",
			in:   LoginInput{Email: "nobody@example.com", Password: "anything"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			wantErr: domain.ErrUnauthorized,
		},
		{
			name: "repo error",
			in:   LoginInput{Email: "u@example.com", Password: "correct-password"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, errDB
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{getByEmailFn: tt.getByEmailFn}
			svc := NewAuthService(userRepo, &mockProfileRepo{}, h, jwtpkg.NewManager("test", time.Hour))

			token, err := svc.Login(context.Background(), tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tt.wantErr)
			}
			if tt.wantToken && token == "" {
				t.Error("expected non-empty token")
			}
			if !tt.wantToken && token != "" {
				t.Errorf("expected empty token, got %q", token)
			}
		})
	}
}
