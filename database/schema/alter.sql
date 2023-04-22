-- alter table statements used to create relationships among tables
-- that are not possibile during the creation (for dependency issues)

-- nullable
ALTER TABLE heart_rate_zones ADD COLUMN IF NOT EXISTS heart_rate_activity_id bigint references heart_rate_activities(id);
