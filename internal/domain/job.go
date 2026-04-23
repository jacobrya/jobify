package domain

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Company     string    `json:"company"`
	Description string    `json:"description"`
	Skills      []string  `json:"skills"`
	SalaryMin   int       `json:"salary_min"`
	SalaryMax   int       `json:"salary_max"`
	IsRemote    bool      `json:"is_remote"`
	Location    string    `json:"location"`
	Source      string    `json:"source"`
	SourceID    string    `json:"source_id,omitempty"`
	URL         string    `json:"url"`
	IsActive    bool      `json:"is_active"`
	ViewsCount  int64     `json:"views_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type JobWithMatch struct {
	Job
	MatchPercent  int      `json:"match_percent"`
	MatchedSkills []string `json:"matched_skills"`
	MissingSkills []string `json:"missing_skills"`
}
