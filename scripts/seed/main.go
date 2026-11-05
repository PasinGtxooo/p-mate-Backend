package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/config"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const schemaSQL = `
-- Extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Project role enum
DO $$ BEGIN
	CREATE TYPE project_role AS ENUM ('owner','pm','member');
EXCEPTION WHEN duplicate_object THEN null; END $$;

-- Accounts
CREATE TABLE IF NOT EXISTS accounts (
	id         CHAR(36) PRIMARY KEY,
	node_id    CHAR(36) NOT NULL,
	username   VARCHAR(100) UNIQUE NOT NULL,
	email      VARCHAR(255) UNIQUE NOT NULL,
	password   VARCHAR(255) NOT NULL,
	state      VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (state IN ('active','inactive')),
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_accounts_node_id ON accounts(node_id);

-- Profiles
CREATE TABLE IF NOT EXISTS profiles (
	id          CHAR(36) PRIMARY KEY,
	node_id     CHAR(36) NOT NULL,
	account_id  CHAR(36) UNIQUE NOT NULL,
	first_name  VARCHAR(100) NOT NULL,
	last_name   VARCHAR(100) NOT NULL,
	nickname    VARCHAR(50),
	avatar_path VARCHAR(500),
	skills      JSONB,
	state       VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (state IN ('active','inactive')),
	created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_profiles_node_id ON profiles(node_id);

-- Projects
CREATE TABLE IF NOT EXISTS projects (
	id         CHAR(36) PRIMARY KEY,
	node_id    CHAR(36) NOT NULL,
	owner_id   CHAR(36) NOT NULL,
	name       VARCHAR(255) NOT NULL,
	config     JSONB,
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_projects_node_id  ON projects(node_id);
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);

-- Project members (composite PK)
CREATE TABLE IF NOT EXISTS project_members (
	project_id CHAR(36) NOT NULL,
	account_id CHAR(36) NOT NULL,
	node_id    CHAR(36) NOT NULL,
	role       project_role NOT NULL DEFAULT 'member',
	created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (project_id, account_id)
);
CREATE INDEX IF NOT EXISTS idx_project_members_node_id ON project_members(node_id);

-- Project invitations
CREATE TABLE IF NOT EXISTS project_invitations (
	id                  UUID PRIMARY KEY,
	node_id             UUID NOT NULL,
	project_id          UUID NOT NULL,
	inviter_account_id  UUID NOT NULL,
	invitee_account_id  UUID NOT NULL,
	role                VARCHAR(50) NOT NULL,
	status              VARCHAR(20) NOT NULL DEFAULT 'pending',
	expires_at          TIMESTAMPTZ NOT NULL,
	responded_at        TIMESTAMPTZ,
	created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at          TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_invitations_node_id ON project_invitations(node_id);
CREATE UNIQUE INDEX IF NOT EXISTS idx_project_invitee ON project_invitations(project_id, invitee_account_id) WHERE deleted_at IS NULL;

-- Tasks
CREATE TABLE IF NOT EXISTS tasks (
	id              CHAR(36) PRIMARY KEY,
	node_id         CHAR(36) NOT NULL,
	project_id      CHAR(36) NOT NULL,
	name            VARCHAR(255) NOT NULL,
	description     TEXT,
	priority        VARCHAR(20) NOT NULL DEFAULT 'medium',
	start_datetime  TIMESTAMPTZ,
	end_datetime    TIMESTAMPTZ,
	location        VARCHAR(255),
	recurring_days  INT,
	recurring_until TIMESTAMPTZ,
	estimate_manday DOUBLE PRECISION,
	required_skill  VARCHAR(100),
	assignee_id     CHAR(36),
	status          VARCHAR(20) NOT NULL DEFAULT 'todo',
	created_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at      TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	deleted_at      TIMESTAMPTZ
);
CREATE INDEX IF NOT EXISTS idx_tasks_node_id     ON tasks(node_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_id  ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee_id ON tasks(assignee_id);
CREATE INDEX IF NOT EXISTS idx_tasks_deleted_at  ON tasks(deleted_at);
`

func main() {
	cfg := config.NewConfig()
	db := database.NewDB(cfg)

	log.Println("📦 creating schema (idempotent) ...")
	if err := db.Exec(schemaSQL).Error; err != nil {
		log.Fatalf("❌ schema create failed: %v", err)
	}
	log.Println("✅ schema OK")

	password := "password123"
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash: %v", err)
	}
	pw := string(hash)
	now := time.Now().UTC()

	nodeID := uuid.New()

	// Three mock users
	users := []struct {
		accountID uuid.UUID
		profileID uuid.UUID
		username  string
		email     string
		firstName string
		lastName  string
		nickname  string
		skills    []map[string]any
	}{
		{uuid.New(), uuid.New(), "alice", "alice@example.com", "Alice", "Wong", "Al",
			[]map[string]any{{"name": "Go", "level": 5}, {"name": "PostgreSQL", "level": 4}}},
		{uuid.New(), uuid.New(), "bob", "bob@example.com", "Bob", "Lee", "Bobby",
			[]map[string]any{{"name": "React", "level": 5}, {"name": "TypeScript", "level": 4}}},
		{uuid.New(), uuid.New(), "carol", "carol@example.com", "Carol", "Chan", "C",
			[]map[string]any{{"name": "Design", "level": 5}, {"name": "Figma", "level": 5}}},
	}

	for _, u := range users {
		if err := db.Exec(`
			INSERT INTO accounts (id, node_id, username, email, password, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
			ON CONFLICT (username) DO NOTHING`,
			u.accountID.String(), nodeID.String(), u.username, u.email, pw, now, now).Error; err != nil {
			log.Fatalf("insert account %s: %v", u.username, err)
		}
		skillsJSON, _ := json.Marshal(u.skills)
		if err := db.Exec(`
			INSERT INTO profiles (id, node_id, account_id, first_name, last_name, nickname, skills, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?::jsonb, 'active', ?, ?)
			ON CONFLICT (account_id) DO NOTHING`,
			u.profileID.String(), nodeID.String(), u.accountID.String(),
			u.firstName, u.lastName, u.nickname, string(skillsJSON), now, now).Error; err != nil {
			log.Fatalf("insert profile %s: %v", u.username, err)
		}
		log.Printf("  👤 %s (%s)", u.username, u.accountID)
	}

	// One demo project owned by alice
	projectID := uuid.New()
	cfgJSON := `{"color":"#4f46e5"}`
	if err := db.Exec(`
		INSERT INTO projects (id, node_id, owner_id, name, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?::jsonb, ?, ?)
		ON CONFLICT (id) DO NOTHING`,
		projectID.String(), nodeID.String(), users[0].accountID.String(),
		"Smart Task Demo", cfgJSON, now, now).Error; err != nil {
		log.Fatalf("insert project: %v", err)
	}
	log.Printf("  📁 project Smart Task Demo (%s)", projectID)

	// Members: alice owner, bob + carol members
	members := []struct {
		accountID uuid.UUID
		role      string
	}{
		{users[0].accountID, "owner"},
		{users[1].accountID, "member"},
		{users[2].accountID, "member"},
	}
	for _, m := range members {
		if err := db.Exec(`
			INSERT INTO project_members (project_id, account_id, node_id, role, created_at, updated_at)
			VALUES (?, ?, ?, ?::project_role, ?, ?)
			ON CONFLICT (project_id, account_id) DO NOTHING`,
			projectID.String(), m.accountID.String(), nodeID.String(), m.role, now, now).Error; err != nil {
			log.Fatalf("insert member: %v", err)
		}
	}

	// Four demo tasks
	tasks := []struct {
		name       string
		desc       string
		priority   string
		status     string
		skill      string
		manday     float64
		assigneeID uuid.UUID
		start      string
		end        string
	}{
		{"Design login screen", "Figma mockup + spec", "high", "in_progress",
			"Design", 1.5, users[2].accountID, "2026-11-05T09:00:00Z", "2026-11-06T18:00:00Z"},
		{"Implement JWT auth", "Login + refresh + middleware", "high", "todo",
			"Go", 2, users[0].accountID, "2026-11-07T09:00:00Z", "2026-11-08T18:00:00Z"},
		{"Build dashboard UI", "Task list + charts", "medium", "todo",
			"React", 3, users[1].accountID, "2026-11-09T09:00:00Z", "2026-11-11T18:00:00Z"},
		{"Write README", "Setup + API docs", "low", "todo",
			"", 0.5, users[0].accountID, "2026-11-12T09:00:00Z", "2026-11-12T18:00:00Z"},
	}
	for _, t := range tasks {
		taskID := uuid.New()
		if err := db.Exec(`
			INSERT INTO tasks (id, node_id, project_id, name, description, priority, status,
				required_skill, estimate_manday, assignee_id, start_datetime, end_datetime,
				created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, NULLIF(?,''), ?, ?, ?, ?, ?, ?)`,
			taskID.String(), nodeID.String(), projectID.String(),
			t.name, t.desc, t.priority, t.status,
			t.skill, t.manday, t.assigneeID.String(), t.start, t.end, now, now).Error; err != nil {
			log.Fatalf("insert task %q: %v", t.name, err)
		}
		log.Printf("  📝 %s", t.name)
	}

	fmt.Println()
	fmt.Println("✅ Seed complete.")
	fmt.Println("   Login creds (all users):  password =", password)
	fmt.Println("   Usernames: alice / bob / carol")
}
