package repository

import (
	"context"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	CreateWithProfile(ctx context.Context, user *domain.User, profile *domain.DeveloperProfile) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
}

type ProfileRepository interface {
	Create(ctx context.Context, profile *domain.DeveloperProfile) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.DeveloperProfile, error)
	Update(ctx context.Context, profile *domain.DeveloperProfile) error
}

type JobFilter struct {
	Skills    []string
	Remote    *bool
	SalaryMin *int
	SalaryMax *int
	Page      int
	Limit     int
}

type JobRepository interface {
	Create(ctx context.Context, job *domain.Job) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	List(ctx context.Context, filter JobFilter) ([]domain.Job, int, error)
	Update(ctx context.Context, job *domain.Job) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsBySourceID(ctx context.Context, sourceID string) (bool, error)
	IncrementViews(ctx context.Context, jobID uuid.UUID) error
}

type ApplicationRepository interface {
	Create(ctx context.Context, app *domain.Application) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Application, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Application, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domain.ApplicationStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type SavedJobRepository interface {
	Create(ctx context.Context, savedJob *domain.SavedJob) error
	Delete(ctx context.Context, userID, jobID uuid.UUID) error
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Job, int, error)
	Exists(ctx context.Context, userID, jobID uuid.UUID) (bool, error)
}

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Duration) error
	GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
