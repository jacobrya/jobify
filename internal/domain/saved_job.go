package domain

import (
	"time"

	"github.com/google/uuid"
)

type SavedJob struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	JobID     uuid.UUID `json:"job_id"`
	CreatedAt time.Time `json:"created_at"`
}
