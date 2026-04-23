package service

import (
	"context"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/google/uuid"
)

type SavedJobService struct {
	repo repository.SavedJobRepository
}

func NewSavedJobService(repo repository.SavedJobRepository) *SavedJobService {
	return &SavedJobService{repo: repo}
}

func (s *SavedJobService) Save(ctx context.Context, userID, jobID uuid.UUID) error {
	exists, err := s.repo.Exists(ctx, userID, jobID)
	if err != nil {
		return err
	}
	if exists {
		return domain.ErrConflict
	}
	return s.repo.Create(ctx, &domain.SavedJob{
		ID:        uuid.New(),
		UserID:    userID,
		JobID:     jobID,
		CreatedAt: time.Now(),
	})
}

func (s *SavedJobService) Unsave(ctx context.Context, userID, jobID uuid.UUID) error {
	return s.repo.Delete(ctx, userID, jobID)
}

func (s *SavedJobService) List(ctx context.Context, userID uuid.UUID, page, limit int) ([]domain.Job, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit
	return s.repo.ListByUser(ctx, userID, limit, offset)
}
