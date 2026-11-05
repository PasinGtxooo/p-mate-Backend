-- Migration: add the 'pm' role to the project_role enum.
-- A PM can manage the project plan (allocate work, edit member skills) like an
-- owner, but is assigned by the owner rather than being the project creator.
--
-- NOTE: ALTER TYPE ... ADD VALUE cannot run inside a transaction block on
-- PostgreSQL older than 12, so run this statement on its own (most migration
-- runners and psql -f handle this fine). IF NOT EXISTS makes it re-runnable.

ALTER TYPE project_role ADD VALUE IF NOT EXISTS 'pm';
