DROP INDEX IF EXISTS idx_jobs_views_count;
ALTER TABLE jobs DROP COLUMN IF EXISTS views_count;
DROP INDEX IF EXISTS idx_saved_jobs_user_id;
DROP TABLE IF EXISTS saved_jobs;