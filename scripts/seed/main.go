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
	position    VARCHAR(50),
	skills      JSONB,
	state       VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (state IN ('active','inactive')),
	created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS position VARCHAR(50);
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

type seedUser struct {
	username  string
	email     string
	firstName string
	lastName  string
	nickname  string
	position  string // PM / Consultant / Developer
	skills    []map[string]any
	accountID uuid.UUID
	profileID uuid.UUID
}

type seedTask struct {
	name       string
	desc       string
	priority   string
	status     string
	skill      string
	manday     float64
	assignee   *seedUser
	phaseStart int // day offset from today
	phaseEnd   int // day offset from today
}

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

	// --- Team: Thai BC (Business Central) consultants/developers ---
	users := []*seedUser{
		{
			username: "pakorn", email: "pakorn@thaibev.co.th",
			firstName: "ปกรณ์", lastName: "ธนากุล", nickname: "โอ",
			position: "PM",
			skills: []map[string]any{
				{"name": "Project Management", "level": 5},
				{"name": "BC Project", "level": 4},
			},
		},
		{
			username: "manee", email: "manee@thaibev.co.th",
			firstName: "มานี", lastName: "ศรีสุข", nickname: "นี",
			position: "Consultant",
			skills: []map[string]any{
				{"name": "BC Finance", "level": 5},
				{"name": "Blueprint", "level": 5},
				{"name": "Fit-Gap", "level": 4},
			},
		},
		{
			username: "somchai", email: "somchai@thaibev.co.th",
			firstName: "สมชาย", lastName: "พานิช", nickname: "ชาย",
			position: "Consultant",
			skills: []map[string]any{
				{"name": "BC Supply Chain", "level": 5},
				{"name": "BC Manufacturing", "level": 4},
			},
		},
		{
			username: "wichai", email: "wichai@thaibev.co.th",
			firstName: "วิชัย", lastName: "ชัยวงศ์", nickname: "ชัย",
			position: "Developer",
			skills: []map[string]any{
				{"name": "AL Development", "level": 5},
				{"name": "Integration", "level": 4},
				{"name": "Extension", "level": 4},
			},
		},
		{
			username: "napha", email: "napha@thaibev.co.th",
			firstName: "นภา", lastName: "วงศ์วาน", nickname: "ภา",
			position: "Developer",
			skills: []map[string]any{
				{"name": "RDLC Report", "level": 5},
				{"name": "Power Platform", "level": 4},
			},
		},
	}

	for _, u := range users {
		u.accountID = uuid.New()
		u.profileID = uuid.New()
		if err := db.Exec(`
			INSERT INTO accounts (id, node_id, username, email, password, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?)
			ON CONFLICT (username) DO NOTHING`,
			u.accountID.String(), nodeID.String(), u.username, u.email, pw, now, now).Error; err != nil {
			log.Fatalf("insert account %s: %v", u.username, err)
		}
		skillsJSON, _ := json.Marshal(u.skills)
		if err := db.Exec(`
			INSERT INTO profiles (id, node_id, account_id, first_name, last_name, nickname, position, skills, state, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?::jsonb, 'active', ?, ?)
			ON CONFLICT (account_id) DO UPDATE SET
				position = EXCLUDED.position,
				skills   = EXCLUDED.skills,
				first_name = EXCLUDED.first_name,
				last_name  = EXCLUDED.last_name`,
			u.profileID.String(), nodeID.String(), u.accountID.String(),
			u.firstName, u.lastName, u.nickname, u.position, string(skillsJSON), now, now).Error; err != nil {
			log.Fatalf("insert profile %s: %v", u.username, err)
		}
		log.Printf("  👤 %s %s — %s", u.firstName, u.lastName, u.position)
	}

	// lookup by position/username for readability
	byName := map[string]*seedUser{}
	for _, u := range users {
		byName[u.username] = u
	}
	pakorn := byName["pakorn"]
	manee := byName["manee"]
	somchai := byName["somchai"]
	wichai := byName["wichai"]
	napha := byName["napha"]

	// --- Project 1: BC Implementation – ไทยเบฟ โลจิสติกส์ (go-live ~3 weeks) ---
	project1ID := uuid.New()
	project1Cfg := `{"color":"#2563eb","go_live":"` + now.AddDate(0, 0, 21).Format("2006-01-02") + `"}`
	if err := db.Exec(`
		INSERT INTO projects (id, node_id, owner_id, name, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?::jsonb, ?, ?)
		ON CONFLICT (id) DO NOTHING`,
		project1ID.String(), nodeID.String(), pakorn.accountID.String(),
		"BC Implementation – ไทยเบฟ โลจิสติกส์", project1Cfg, now, now).Error; err != nil {
		log.Fatalf("insert project1: %v", err)
	}
	log.Printf("  📁 โปรเจกต์: BC Implementation – ไทยเบฟ โลจิสติกส์ (go-live ~3 สัปดาห์)")

	// --- Project 2: BC Phase 2 – ไทยเบฟ (same go-live, go-live conflict narrative) ---
	project2ID := uuid.New()
	project2Cfg := `{"color":"#f59e0b","go_live":"` + now.AddDate(0, 0, 21).Format("2006-01-02") + `"}`
	if err := db.Exec(`
		INSERT INTO projects (id, node_id, owner_id, name, config, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?::jsonb, ?, ?)
		ON CONFLICT (id) DO NOTHING`,
		project2ID.String(), nodeID.String(), pakorn.accountID.String(),
		"BC Phase 2 – ไทยเบฟ", project2Cfg, now, now).Error; err != nil {
		log.Fatalf("insert project2: %v", err)
	}
	log.Printf("  📁 โปรเจกต์: BC Phase 2 – ไทยเบฟ (go-live วันเดียวกัน → Conflict)")

	// Members: pakorn owner on both, others member
	addMember := func(projID uuid.UUID, u *seedUser, role string) {
		if err := db.Exec(`
			INSERT INTO project_members (project_id, account_id, node_id, role, created_at, updated_at)
			VALUES (?, ?, ?, ?::project_role, ?, ?)
			ON CONFLICT (project_id, account_id) DO UPDATE SET role = EXCLUDED.role`,
			projID.String(), u.accountID.String(), nodeID.String(), role, now, now).Error; err != nil {
			log.Fatalf("insert member: %v", err)
		}
	}
	for _, p := range []uuid.UUID{project1ID, project2ID} {
		addMember(p, pakorn, "owner")
		addMember(p, manee, "member")
		addMember(p, somchai, "member")
		addMember(p, wichai, "member")
		addMember(p, napha, "member")
	}

	// --- Project 1 tasks (11 รายการตาม phase) ---
	// Phase windows (day offset from today):
	//   Plan     days 0-1
	//   Analysis days 2-8
	//   Develop  days 9-15
	//   Deploy   days 16-20
	//   Go-live  day 21
	tasks := []seedTask{
		{"Kick-off meeting", "ประชุมเปิดโครงการกับไทยเบฟ ชี้แจง scope และ timeline", "high", "done",
			"Project Management", 1, pakorn, 0, 1},
		{"เก็บ Requirement / Blueprint", "สัมภาษณ์ผู้ใช้งานฝ่ายบัญชี-การเงิน และเขียน Blueprint", "high", "in_progress",
			"BC Finance", 5, manee, 2, 6},
		{"Fit-Gap Analysis", "เทียบ Requirement vs ความสามารถ standard ของ BC", "high", "todo",
			"Fit-Gap", 3, manee, 6, 8},
		{"Config BC Finance", "ตั้งค่า G/L, Posting Group, Dimension", "high", "todo",
			"BC Finance", 4, manee, 9, 12},
		{"Config Supply Chain", "ตั้งค่า Item, Warehouse, Transfer", "high", "todo",
			"BC Supply Chain", 4, somchai, 9, 12},
		{"พัฒนา AL Custom Invoice", "Customize หัว Invoice ตาม format กรมสรรพากร", "high", "todo",
			"AL Development", 6, wichai, 9, 14},
		{"สร้าง Integration API", "สร้าง REST API เชื่อม POS → BC", "high", "todo",
			"Integration", 4, wichai, 12, 15},
		{"พัฒนา RDLC Report", "รายงานขายรายวัน + stock report (ไทย/EN)", "medium", "todo",
			"RDLC Report", 2, napha, 11, 13},
		{"Migrate Master Data", "นำเข้า Chart of Accounts, Items, Customers, Vendors", "high", "todo",
			"", 3, somchai, 16, 18},
		{"UAT กับลูกค้า", "User Acceptance Test ทุก module กับฝ่ายปฏิบัติ", "high", "todo",
			"", 3, manee, 18, 20},
		{"Go-live Support (standby)", "เตรียมพร้อม standby ช่วย onsite วัน go-live", "urgent", "todo",
			"", 2, pakorn, 21, 21},
	}

	for _, t := range tasks {
		taskID := uuid.New()
		start := now.AddDate(0, 0, t.phaseStart).Add(2 * time.Hour) // start 09:00 UTC+7 ≈ 02:00 UTC
		end := now.AddDate(0, 0, t.phaseEnd).Add(11 * time.Hour)    // end 18:00 UTC+7 ≈ 11:00 UTC
		var assignee *string
		if t.assignee != nil {
			id := t.assignee.accountID.String()
			assignee = &id
		}
		var skill *string
		if t.skill != "" {
			s := t.skill
			skill = &s
		}
		if err := db.Exec(`
			INSERT INTO tasks (id, node_id, project_id, name, description, priority, status,
				required_skill, estimate_manday, assignee_id, start_datetime, end_datetime,
				created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			taskID.String(), nodeID.String(), project1ID.String(),
			t.name, t.desc, t.priority, t.status,
			skill, t.manday, assignee, start, end, now, now).Error; err != nil {
			log.Fatalf("insert task %q: %v", t.name, err)
		}
		marker := " "
		if t.assignee == wichai {
			marker = "⚠" // flag workload-heavy dev
		}
		log.Printf("  %s 📝 %s — %s (%.1f MD)", marker, t.name,
			map[bool]string{true: t.assignee.firstName, false: "-"}[t.assignee != nil], t.manday)
	}

	fmt.Println()
	fmt.Println("✅ Seed complete.")
	fmt.Println("   Login creds (all users):  password =", password)
	fmt.Println("   Usernames: pakorn (PM) / manee / somchai / wichai / napha")
	fmt.Println()
	fmt.Println("   Workload note: วิชัย ได้รับงาน 6+4 = 10 MD > deadline window 8")
	fmt.Println("                  → manpower view ควรแสดงสีแดง (over-capacity)")
}
