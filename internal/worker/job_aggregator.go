package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/google/uuid"
)

type JobAggregator struct {
	jobRepo    repository.JobRepository
	logger     *slog.Logger
	ticker     *time.Ticker
	apiURL     string
	httpClient *http.Client
}

func NewJobAggregator(jobRepo repository.JobRepository, logger *slog.Logger, interval time.Duration, apiURL string) *JobAggregator {
	return &JobAggregator{
		jobRepo:    jobRepo,
		logger:     logger,
		ticker:     time.NewTicker(interval),
		apiURL:     apiURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

func (w *JobAggregator) Start(ctx context.Context) {
	w.logger.Info("job aggregator started")
	w.run(ctx)

	for {
		select {
		case <-w.ticker.C:
			w.run(ctx)
		case <-ctx.Done():
			w.ticker.Stop()
			w.logger.Info("job aggregator stopped")
			return
		}
	}
}

func (w *JobAggregator) run(ctx context.Context) {
	jobs, err := w.fetchFromRemotive(ctx)
	if err != nil {
		w.logger.Error("failed to fetch jobs from remotive", "err", err)
		return
	}

	saved := 0
	for _, job := range jobs {
		exists, err := w.jobRepo.ExistsBySourceID(ctx, job.SourceID)
		if err != nil {
			w.logger.Error("failed to check source_id", "err", err)
			continue
		}
		if !exists {
			if err := w.jobRepo.Create(ctx, &job); err != nil {
				w.logger.Error("failed to save job", "err", err)
				continue
			}
			saved++
		}
	}
	w.logger.Info("aggregator run complete", "saved", saved, "total", len(jobs))
}

type remotiveResponse struct {
	Jobs []remotiveJob `json:"jobs"`
}

type remotiveJob struct {
	ID          int      `json:"id"`
	Title       string   `json:"title"`
	CompanyName string   `json:"company_name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	JobType     string   `json:"job_type"`
	URL         string   `json:"url"`
	CandidateRequiredLocation string `json:"candidate_required_location"`
}

func (w *JobAggregator) fetchFromRemotive(ctx context.Context) ([]domain.Job, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, w.apiURL+"?category=software-dev", nil)
	if err != nil {
		return nil, err
	}

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("remotive API returned status %d", resp.StatusCode)
	}

	var remotiveResp remotiveResponse
	if err := json.NewDecoder(resp.Body).Decode(&remotiveResp); err != nil {
		return nil, err
	}

	jobs := make([]domain.Job, 0, len(remotiveResp.Jobs))
	for _, rj := range remotiveResp.Jobs {
		jobs = append(jobs, domain.Job{
			ID:          uuid.New(),
			Title:       rj.Title,
			Company:     rj.CompanyName,
			Description: rj.Description,
			Skills:      rj.Tags,
			IsRemote:    true, // remotive is a remote-only board
			Location:    rj.CandidateRequiredLocation,
			JobType:     rj.JobType,
			Source:      "remotive",
			SourceID:    fmt.Sprintf("remotive_%d", rj.ID),
			URL:         rj.URL,
			IsActive:    true,
		})
	}
	return jobs, nil
}
