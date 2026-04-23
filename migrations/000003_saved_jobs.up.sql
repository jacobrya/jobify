CREATE TABLE saved_jobs (
                            id UUID PRIMARY KEY,
                            user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                            job_id UUID NOT NULL REFERENCES jobs(id) ON DELETE CASCADE,
                            created_at TIMESTAMPTZ DEFAULT NOW(),
                            UNIQUE(user_id, job_id)
);

CREATE INDEX idx_saved_jobs_user_id ON saved_jobs(user_id);

ALTER TABLE jobs ADD COLUMN views_count BIGINT NOT NULL DEFAULT 0;

CREATE INDEX idx_jobs_views_count ON jobs(views_count DESC);