package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/abzalserikbay/jobify/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type JobRepo struct {
	db *pgxpool.Pool
}

func NewJobRepo(db *pgxpool.Pool) *JobRepo {
	return &JobRepo{db: db}
}

func (r *JobRepo) Create(ctx context.Context, j *domain.Job) error {
	q := `INSERT INTO jobs (id, title, company, description, skills, salary_min, salary_max, is_remote, location, job_type, source, source_id, url, is_active)
	      VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	      RETURNING created_at`
	return r.db.QueryRow(ctx, q,
		j.ID, j.Title, j.Company, j.Description, j.Skills,
		j.SalaryMin, j.SalaryMax, j.IsRemote, j.Location, j.JobType,
		j.Source, j.SourceID, j.URL, j.IsActive,
	).Scan(&j.CreatedAt)
}

func (r *JobRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Job, error) {
	var j domain.Job
	q := `SELECT id, title, company, description, skills, salary_min, salary_max,
	             is_remote, location, job_type, source, source_id, url, is_active, views_count, created_at
	      FROM jobs WHERE id = $1 AND is_active = true`
	err := r.db.QueryRow(ctx, q, id).Scan(
		&j.ID, &j.Title, &j.Company, &j.Description, &j.Skills,
		&j.SalaryMin, &j.SalaryMax, &j.IsRemote, &j.Location, &j.JobType,
		&j.Source, &j.SourceID, &j.URL, &j.IsActive, &j.ViewsCount, &j.CreatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &j, err
}

func (r *JobRepo) List(ctx context.Context, f repository.JobFilter) ([]domain.Job, int, error) {
	where := []string{"is_active = true"}
	args := []any{}
	i := 1

	if len(f.Skills) > 0 {
		where = append(where, fmt.Sprintf("skills && $%d", i))
		args = append(args, f.Skills)
		i++
	}
	if f.Remote != nil {
		where = append(where, fmt.Sprintf("is_remote = $%d", i))
		args = append(args, *f.Remote)
		i++
	}
	if f.SalaryMin != nil {
		where = append(where, fmt.Sprintf("salary_max >= $%d", i))
		args = append(args, *f.SalaryMin)
		i++
	}
	if f.SalaryMax != nil {
		where = append(where, fmt.Sprintf("salary_min <= $%d", i))
		args = append(args, *f.SalaryMax)
		i++
	}

	cond := strings.Join(where, " AND ")

	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM jobs WHERE "+cond, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if f.Limit == 0 {
		f.Limit = 20
	}
	if f.Page == 0 {
		f.Page = 1
	}
	offset := (f.Page - 1) * f.Limit

	args = append(args, f.Limit, offset)
	q := fmt.Sprintf(`SELECT id, title, company, description, skills, salary_min, salary_max,
	                         is_remote, location, job_type, source, source_id, url, is_active, views_count, created_at
	                  FROM jobs WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, cond, i, i+1)

	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var jobs []domain.Job
	for rows.Next() {
		var j domain.Job
		if err := rows.Scan(
			&j.ID, &j.Title, &j.Company, &j.Description, &j.Skills,
			&j.SalaryMin, &j.SalaryMax, &j.IsRemote, &j.Location, &j.JobType,
			&j.Source, &j.SourceID, &j.URL, &j.IsActive, &j.ViewsCount, &j.CreatedAt,
		); err != nil {
			return nil, 0, err
		}
		jobs = append(jobs, j)
	}
	return jobs, total, nil
}

func (r *JobRepo) Update(ctx context.Context, j *domain.Job) error {
	q := `UPDATE jobs SET title=$1, company=$2, description=$3, skills=$4, salary_min=$5, salary_max=$6,
	      is_remote=$7, location=$8, job_type=$9, url=$10, is_active=$11 WHERE id=$12`
	_, err := r.db.Exec(ctx, q,
		j.Title, j.Company, j.Description, j.Skills,
		j.SalaryMin, j.SalaryMax, j.IsRemote, j.Location, j.JobType, j.URL, j.IsActive, j.ID,
	)
	return err
}

func (r *JobRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE jobs SET is_active=false WHERE id=$1`, id)
	return err
}

func (r *JobRepo) ExistsBySourceID(ctx context.Context, sourceID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM jobs WHERE source_id=$1)`, sourceID).Scan(&exists)
	return exists, err
}

func (r *JobRepo) IncrementViews(ctx context.Context, jobID uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE jobs SET views_count = views_count + 1 WHERE id = $1`, jobID)
	if err != nil {
		return fmt.Errorf("job_repo increment_views: %w", err)
	}
	return nil
}
