package postgres

import (
	"context"
	"errors"

	"github.com/abzalserikbay/jobify/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepo struct {
	db *pgxpool.Pool
}

func NewUserRepo(db *pgxpool.Pool) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(ctx context.Context, u *domain.User) error {
	q := `INSERT INTO users (id, email, password, role)
	      VALUES ($1, $2, $3, $4)
	      RETURNING created_at, updated_at`
	return r.db.QueryRow(ctx, q, u.ID, u.Email, u.Password, u.Role).
		Scan(&u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	var u domain.User
	q := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRow(ctx, q, id).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}

func (r *UserRepo) CreateWithProfile(ctx context.Context, u *domain.User, p *domain.DeveloperProfile) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := `INSERT INTO users (id, email, password, role) VALUES ($1, $2, $3, $4) RETURNING created_at, updated_at`
	if err := tx.QueryRow(ctx, q, u.ID, u.Email, u.Password, u.Role).Scan(&u.CreatedAt, &u.UpdatedAt); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO developer_profiles (id, user_id) VALUES ($1, $2)`, p.ID, p.UserID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	q := `SELECT id, email, password, role, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRow(ctx, q, email).
		Scan(&u.ID, &u.Email, &u.Password, &u.Role, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrNotFound
	}
	return &u, err
}
