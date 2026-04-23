package service

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/google/uuid"
)

func TestCalculateMatch(t *testing.T) {
	tests := []struct {
		name        string
		user        []string
		job         []string
		wantPercent int
		wantMatched []string
		wantMissing []string
	}{
		{
			name:        "perfect match",
			user:        []string{"Go", "Docker"},
			job:         []string{"Go", "Docker"},
			wantPercent: 100,
			wantMatched: []string{"docker", "go"},
			wantMissing: nil,
		},
		{
			name:        "half match",
			user:        []string{"Go"},
			job:         []string{"Go", "Java"},
			wantPercent: 50,
			wantMatched: []string{"go"},
			wantMissing: []string{"java"},
		},
		{
			name:        "no match",
			user:        []string{"PHP"},
			job:         []string{"Rust"},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: []string{"rust"},
		},
		{
			name:        "empty user skills",
			user:        []string{},
			job:         []string{"Go"},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: []string{"go"},
		},
		{
			name:        "case insensitive",
			user:        []string{"go", "DOCKER"},
			job:         []string{"Go", "docker"},
			wantPercent: 100,
			wantMatched: []string{"docker", "go"},
			wantMissing: nil,
		},
		{
			name:        "both empty",
			user:        []string{},
			job:         []string{},
			wantPercent: 0,
			wantMatched: nil,
			wantMissing: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			percent, matched, missing := CalculateMatch(tt.user, tt.job)
			if percent != tt.wantPercent {
				t.Errorf("percent: got %d, want %d", percent, tt.wantPercent)
			}
			sort.Strings(matched)
			sort.Strings(missing)
			if !equalSlices(matched, tt.wantMatched) {
				t.Errorf("matched: got %v, want %v", matched, tt.wantMatched)
			}
			if !equalSlices(missing, tt.wantMissing) {
				t.Errorf("missing: got %v, want %v", missing, tt.wantMissing)
			}
		})
	}
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type mockJobRepo struct {
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*domain.Job, error)
	incrementViewFn func(ctx context.Context, id uuid.UUID) error
}

func (m *mockJobRepo) Create(_ context.Context, _ *domain.Job) error { return nil }
func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockJobRepo) List(_ context.Context, _ repository.JobFilter) ([]domain.Job, int, error) {
	return nil, 0, nil
}
func (m *mockJobRepo) Update(_ context.Context, _ *domain.Job) error { return nil }
func (m *mockJobRepo) Delete(_ context.Context, _ uuid.UUID) error   { return nil }
func (m *mockJobRepo) ExistsBySourceID(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockJobRepo) IncrementViews(ctx context.Context, id uuid.UUID) error {
	return m.incrementViewFn(ctx, id)
}

func TestJobService_GetByID_IncrementsViews(t *testing.T) {
	jobID := uuid.New()
	called := make(chan struct{}, 1)

	repo := &mockJobRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Job, error) {
			return &domain.Job{ID: jobID, Skills: []string{"Go"}}, nil
		},
		incrementViewFn: func(_ context.Context, id uuid.UUID) error {
			if id == jobID {
				called <- struct{}{}
			}
			return nil
		},
	}

	svc := &JobService{jobRepo: repo, cache: nil}
	_, err := svc.GetByID(context.Background(), jobID, []string{"Go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case <-called:
	case <-time.After(500 * time.Millisecond):
		t.Error("IncrementViews was not called")
	}
}
