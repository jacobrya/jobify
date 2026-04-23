package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SavedJobRepo struct {
	db *pgxpool.Pool
}

func NewSavedJobRepo(db *pgxpool.Pool) *SavedJobRepo {
	return &SavedJobRepo{db: db}
}

func (r *SavedJobRepo) Create(ctx context.Context, s *domain.SavedJob) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO saved_jobs (id, user_id, job_id, created_at) VALUES ($1, $2, $3, $4)`,
		s.ID, s.UserID, s.JobID, s.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("saved_job_repo create: %w", err)
	}
	return nil
}

func (r *SavedJobRepo) Delete(ctx context.Context, userID, jobID uuid.UUID) error {
	tag, err := r.db.Exec(ctx,
		`DELETE FROM saved_jobs WHERE user_id = $1 AND job_id = $2`,
		userID, jobID,
	)
	if err != nil {
		return fmt.Errorf("saved_job_repo delete: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *SavedJobRepo) Exists(ctx context.Context, userID, jobID uuid.UUID) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM saved_jobs WHERE user_id = $1 AND job_id = $2)`,
		userID, jobID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("saved_job_repo exists: %w", err)
	}
	return exists, nil
}

func (r *SavedJobRepo) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.Job, int, error) {
	const q = `
		SELECT j.id, j.title, j.company, j.description, j.skills,
		       j.salary_min, j.salary_max, j.is_remote, j.location,
		       j.source, j.source_id, j.url, j.is_active, j.views_count, j.created_at
		FROM saved_jobs sj
		JOIN jobs j ON j.id = sj.job_id
		WHERE sj.user_id = $1
		ORDER BY sj.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, q, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("saved_job_repo list: %w", err)
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(
			&j.ID, &j.Title, &j.Company, &j.Description, &j.Skills,
			&j.SalaryMin, &j.SalaryMax, &j.IsRemote, &j.Location,
			&j.Source, &j.SourceID, &j.URL, &j.IsActive, &j.ViewsCount, &j.CreatedAt,
		); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, 0, domain.ErrNotFound
			}
			return nil, 0, fmt.Errorf("saved_job_repo list scan: %w", err)
		}
		jobs = append(jobs, j)
	}

	var total int
	err = r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM saved_jobs WHERE user_id = $1`, userID,
	).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("saved_job_repo list count: %w", err)
	}

	return jobs, total, nil
}
