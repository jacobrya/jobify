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

// mockUserRepo implements repository.UserRepository.
type mockUserRepo struct {
	getByEmailFn        func(ctx context.Context, email string) (*domain.User, error)
	getByIDFn           func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	createFn            func(ctx context.Context, u *domain.User) error
	createWithProfileFn func(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error
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
func (m *mockUserRepo) CreateWithProfile(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error {
	return m.createWithProfileFn(ctx, u, p)
}

// mockTokenRepo implements repository.TokenRepository with an in-memory store.
type mockTokenRepo struct {
	tokens   map[string]uuid.UUID
	storeErr error
}

func newMockTokenRepo() *mockTokenRepo {
	return &mockTokenRepo{tokens: map[string]uuid.UUID{}}
}

func (m *mockTokenRepo) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Duration) error {
	if m.storeErr != nil {
		return m.storeErr
	}
	m.tokens[token] = userID
	return nil
}
func (m *mockTokenRepo) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	id, ok := m.tokens[token]
	if !ok {
		return uuid.Nil, domain.ErrNotFound
	}
	return id, nil
}
func (m *mockTokenRepo) DeleteRefreshToken(ctx context.Context, token string) error {
	delete(m.tokens, token)
	return nil
}

func TestAuthService_Register(t *testing.T) {
	errDB := errors.New("db down")

	tests := []struct {
		name                string
		in                  RegisterInput
		getByEmailFn        func(ctx context.Context, email string) (*domain.User, error)
		createWithProfileFn func(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error
		wantErr             error
	}{
		{
			name: "success",
			in:   RegisterInput{Email: "new@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			createWithProfileFn: func(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error {
				return nil
			},
			wantErr: nil,
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
			name: "create tx error",
			in:   RegisterInput{Email: "y@example.com", Password: "secret123"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return nil, domain.ErrNotFound
			},
			createWithProfileFn: func(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error {
				return errDB
			},
			wantErr: errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{
				getByEmailFn:        tt.getByEmailFn,
				createWithProfileFn: tt.createWithProfileFn,
			}
			svc := NewAuthService(userRepo, newMockTokenRepo(), hasher.New(), jwtpkg.NewManager("test", time.Hour), 7*24*time.Hour)

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
					t.Error("password must be hashed")
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
		tokenStoreErr error
		wantErr      error
		wantPair     bool
	}{
		{
			name: "success",
			in:   LoginInput{Email: "u@example.com", Password: "correct-password"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
			},
			wantErr:  nil,
			wantPair: true,
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
		{
			name: "token store error",
			in:   LoginInput{Email: "u@example.com", Password: "correct-password"},
			getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
				return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
			},
			tokenStoreErr: errDB,
			wantErr:       errDB,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userRepo := &mockUserRepo{getByEmailFn: tt.getByEmailFn}
			tokenRepo := &mockTokenRepo{tokens: map[string]uuid.UUID{}, storeErr: tt.tokenStoreErr}
			svc := NewAuthService(userRepo, tokenRepo, h, jwtpkg.NewManager("test", time.Hour), 7*24*time.Hour)

			pair, err := svc.Login(context.Background(), tt.in)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err: got %v, want %v", err, tt.wantErr)
			}
			if tt.wantPair {
				if pair == nil {
					t.Fatal("expected token pair, got nil")
				}
				if pair.AccessToken == "" {
					t.Error("expected non-empty access token")
				}
				if pair.RefreshToken == "" {
					t.Error("expected non-empty refresh token")
				}
			}
			if !tt.wantPair && pair != nil {
				t.Errorf("expected nil pair, got %+v", pair)
			}
		})
	}
}

func TestAuthService_Refresh(t *testing.T) {
	h := hasher.New()
	correctHash, _ := h.Hash("pass")
	uid := uuid.New()

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
		},
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*domain.User, error) {
			if id != uid {
				return nil, domain.ErrNotFound
			}
			return &domain.User{ID: uid, Email: "u@example.com", Password: correctHash, Role: domain.RoleDeveloper}, nil
		},
	}

	tokenRepo := newMockTokenRepo()
	svc := NewAuthService(userRepo, tokenRepo, h, jwtpkg.NewManager("test", time.Hour), 7*24*time.Hour)

	// login to get initial pair
	pair, err := svc.Login(context.Background(), LoginInput{Email: "u@example.com", Password: "pass"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	// refresh
	newPair, err := svc.Refresh(context.Background(), pair.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if newPair.AccessToken == "" || newPair.RefreshToken == "" {
		t.Error("expected non-empty tokens in new pair")
	}

	// old refresh token must be invalidated
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized for used token, got %v", err)
	}
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	svc := NewAuthService(
		&mockUserRepo{},
		newMockTokenRepo(),
		hasher.New(),
		jwtpkg.NewManager("test", time.Hour),
		time.Hour,
	)
	_, err := svc.Refresh(context.Background(), "bogus-token")
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized, got %v", err)
	}
}

func TestAuthService_Logout(t *testing.T) {
	h := hasher.New()
	correctHash, _ := h.Hash("pass")
	uid := uuid.New()

	userRepo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, email string) (*domain.User, error) {
			return &domain.User{ID: uid, Email: email, Password: correctHash, Role: domain.RoleDeveloper}, nil
		},
	}
	tokenRepo := newMockTokenRepo()
	svc := NewAuthService(userRepo, tokenRepo, h, jwtpkg.NewManager("test", time.Hour), time.Hour)

	pair, err := svc.Login(context.Background(), LoginInput{Email: "u@example.com", Password: "pass"})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}

	if err := svc.Logout(context.Background(), pair.RefreshToken); err != nil {
		t.Fatalf("Logout: %v", err)
	}

	// token must be gone — refresh should fail
	_, err = svc.Refresh(context.Background(), pair.RefreshToken)
	if !errors.Is(err, domain.ErrUnauthorized) {
		t.Errorf("expected ErrUnauthorized after logout, got %v", err)
	}
}
