package worker

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/google/uuid"
)

type mockJobRepo struct {
	existsBySourceIDFn func(ctx context.Context, sourceID string) (bool, error)
	createFn           func(ctx context.Context, job *domain.Job) error
	created            []*domain.Job
}

func (m *mockJobRepo) ExistsBySourceID(ctx context.Context, sourceID string) (bool, error) {
	return m.existsBySourceIDFn(ctx, sourceID)
}
func (m *mockJobRepo) Create(ctx context.Context, job *domain.Job) error {
	m.created = append(m.created, job)
	if m.createFn != nil {
		return m.createFn(ctx, job)
	}
	return nil
}
func (m *mockJobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	return nil, nil
}
func (m *mockJobRepo) List(ctx context.Context, f repository.JobFilter) ([]domain.Job, int, error) {
	return nil, 0, nil
}
func (m *mockJobRepo) Update(ctx context.Context, job *domain.Job) error      { return nil }
func (m *mockJobRepo) Delete(ctx context.Context, id uuid.UUID) error         { return nil }
func (m *mockJobRepo) IncrementViews(ctx context.Context, id uuid.UUID) error { return nil }

func newTestAggregator(repoArg *mockJobRepo, url string) *JobAggregator {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	return NewJobAggregator(repoArg, logger, time.Hour, url)
}

func TestJobAggregator_FetchAndSave(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := remotiveResponse{
			Jobs: []remotiveJob{
				{
					ID:                        1,
					Title:                     "Go Engineer",
					CompanyName:               "Acme",
					Description:               "Build cool stuff",
					Tags:                      []string{"go", "docker"},
					JobType:                   "full_time",
					URL:                       "https://example.com/job/1",
					CandidateRequiredLocation: "Worldwide",
				},
				{
					ID:                        2,
					Title:                     "Rust Dev",
					CompanyName:               "Widgets",
					Description:               "Write safe code",
					Tags:                      []string{"rust"},
					JobType:                   "contract",
					URL:                       "https://example.com/job/2",
					CandidateRequiredLocation: "USA",
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	repo := &mockJobRepo{
		existsBySourceIDFn: func(ctx context.Context, sourceID string) (bool, error) {
			return false, nil
		},
	}
	agg := newTestAggregator(repo, srv.URL)
	agg.run(context.Background())

	if len(repo.created) != 2 {
		t.Fatalf("expected 2 saved jobs, got %d", len(repo.created))
	}

	j := repo.created[0]
	if j.Title != "Go Engineer" {
		t.Errorf("title: got %q, want %q", j.Title, "Go Engineer")
	}
	if j.JobType != "full_time" {
		t.Errorf("job_type: got %q, want %q", j.JobType, "full_time")
	}
	if !j.IsRemote {
		t.Error("expected IsRemote=true")
	}
	if j.Source != "remotive" {
		t.Errorf("source: got %q, want %q", j.Source, "remotive")
	}
	if j.SourceID != "remotive_1" {
		t.Errorf("source_id: got %q, want %q", j.SourceID, "remotive_1")
	}
	if j.ID == uuid.Nil {
		t.Error("expected non-nil UUID")
	}

	j2 := repo.created[1]
	if j2.JobType != "contract" {
		t.Errorf("job_type[1]: got %q, want %q", j2.JobType, "contract")
	}
}

func TestJobAggregator_SkipsExisting(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := remotiveResponse{
			Jobs: []remotiveJob{
				{ID: 10, Title: "Old Job", CompanyName: "Corp", URL: "https://x.com"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	repo := &mockJobRepo{
		existsBySourceIDFn: func(ctx context.Context, sourceID string) (bool, error) {
			return true, nil
		},
	}
	newTestAggregator(repo, srv.URL).run(context.Background())

	if len(repo.created) != 0 {
		t.Errorf("expected no new jobs, got %d", len(repo.created))
	}
}

func TestJobAggregator_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	repo := &mockJobRepo{
		existsBySourceIDFn: func(ctx context.Context, sourceID string) (bool, error) {
			return false, nil
		},
	}
	newTestAggregator(repo, srv.URL).run(context.Background())

	if len(repo.created) != 0 {
		t.Errorf("expected no saved jobs on API error, got %d", len(repo.created))
	}
}

func TestJobAggregator_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(remotiveResponse{Jobs: []remotiveJob{}})
	}))
	defer srv.Close()

	repo := &mockJobRepo{
		existsBySourceIDFn: func(ctx context.Context, sourceID string) (bool, error) {
			return false, nil
		},
	}
	newTestAggregator(repo, srv.URL).run(context.Background())

	if len(repo.created) != 0 {
		t.Errorf("expected 0 jobs for empty response, got %d", len(repo.created))
	}
}
