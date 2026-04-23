package service

import (
	"context"
	"testing"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
)

type mockSavedJobRepo struct {
	createFn     func(ctx context.Context, s *domain.SavedJob) error
	deleteFn     func(ctx context.Context, userID, jobID uuid.UUID) error
	listByUserFn func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Job, int, error)
	existsFn     func(ctx context.Context, userID, jobID uuid.UUID) (bool, error)
}

func (m *mockSavedJobRepo) Create(ctx context.Context, s *domain.SavedJob) error {
	return m.createFn(ctx, s)
}
func (m *mockSavedJobRepo) Delete(ctx context.Context, userID, jobID uuid.UUID) error {
	return m.deleteFn(ctx, userID, jobID)
}
func (m *mockSavedJobRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Job, int, error) {
	return m.listByUserFn(ctx, userID, limit, offset)
}
func (m *mockSavedJobRepo) Exists(ctx context.Context, userID, jobID uuid.UUID) (bool, error) {
	return m.existsFn(ctx, userID, jobID)
}

func TestSavedJobService_Save(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockSavedJobRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockSavedJobRepo{
				existsFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) { return false, nil },
				createFn: func(_ context.Context, _ *domain.SavedJob) error { return nil },
			},
			wantErr: nil,
		},
		{
			name: "already saved",
			repo: &mockSavedJobRepo{
				existsFn: func(_ context.Context, _, _ uuid.UUID) (bool, error) { return true, nil },
			},
			wantErr: domain.ErrConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSavedJobService(tt.repo)
			err := svc.Save(context.Background(), userID, jobID)
			if err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSavedJobService_Unsave(t *testing.T) {
	userID := uuid.New()
	jobID := uuid.New()

	tests := []struct {
		name    string
		repo    *mockSavedJobRepo
		wantErr error
	}{
		{
			name: "success",
			repo: &mockSavedJobRepo{
				deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return nil },
			},
			wantErr: nil,
		},
		{
			name: "not found",
			repo: &mockSavedJobRepo{
				deleteFn: func(_ context.Context, _, _ uuid.UUID) error { return domain.ErrNotFound },
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewSavedJobService(tt.repo)
			err := svc.Unsave(context.Background(), userID, jobID)
			if err != tt.wantErr {
				t.Errorf("got %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestSavedJobService_List(t *testing.T) {
	userID := uuid.New()
	jobs := []domain.Job{
		{ID: uuid.New(), Title: "Go Dev", CreatedAt: time.Now()},
		{ID: uuid.New(), Title: "Rust Dev", CreatedAt: time.Now()},
	}

	tests := []struct {
		name      string
		page      int
		limit     int
		wantLimit int
		wantOff   int
	}{
		{name: "default pagination", page: 0, limit: 0, wantLimit: 20, wantOff: 0},
		{name: "page 2 limit 5", page: 2, limit: 5, wantLimit: 5, wantOff: 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSavedJobRepo{
				listByUserFn: func(_ context.Context, _ uuid.UUID, limit, offset int) ([]domain.Job, int, error) {
					if limit != tt.wantLimit {
						t.Errorf("limit: got %d, want %d", limit, tt.wantLimit)
					}
					if offset != tt.wantOff {
						t.Errorf("offset: got %d, want %d", offset, tt.wantOff)
					}
					return jobs, len(jobs), nil
				},
			}
			svc := NewSavedJobService(repo)
			got, total, err := svc.List(context.Background(), userID, tt.page, tt.limit)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if total != len(jobs) {
				t.Errorf("total: got %d, want %d", total, len(jobs))
			}
			if len(got) != len(jobs) {
				t.Errorf("jobs count: got %d, want %d", len(got), len(jobs))
			}
		})
	}
}
