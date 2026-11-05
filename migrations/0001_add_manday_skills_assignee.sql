-- Migration: add manday, required skill, assignee to tasks; skills to profiles.
-- Supports the teammate-matching + manpower features.
-- Safe to run multiple times (IF NOT EXISTS guards).

-- Tasks: effort estimate (in manday), skill the task needs, and who it's assigned to.
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS estimate_manday DOUBLE PRECISION;
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS required_skill  VARCHAR(100);
ALTER TABLE tasks ADD COLUMN IF NOT EXISTS assignee_id     CHAR(36);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks (assignee_id);

-- Profiles: skills as JSON array of {name, level}. Stored as JSONB.
-- GORM serializer:json writes/reads JSON text; JSONB accepts it and is queryable.
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS skills JSONB;
