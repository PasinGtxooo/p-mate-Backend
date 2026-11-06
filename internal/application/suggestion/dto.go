package suggestion

import profileEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"

// SuggestTeammateResponse is the ranked list of teammates for a task.
type SuggestTeammateResponse struct {
	TaskID         string              `json:"task_id"`
	TaskName       string              `json:"task_name"`
	RequiredSkill  string              `json:"required_skill,omitempty"`
	EstimateManday float64             `json:"estimate_manday"`
	Candidates     []TeammateCandidate `json:"candidates"`
}

// TeammateCandidate is one teammate scored against the task.
type TeammateCandidate struct {
	AccountID   string  `json:"account_id"`
	Name        string  `json:"name"`
	SkillLevel  int     `json:"skill_level"`
	SkillMatch  int     `json:"skill_match"`
	AssignedMD  float64 `json:"assigned_md"`
	FreeMD      float64 `json:"free_md"`
	Fits        bool    `json:"fits"`
	Recommended bool    `json:"recommended"`
}

// ManpowerResponse tells the PM whether the team can absorb all open work.
type ManpowerResponse struct {
	ProjectID        string       `json:"project_id"`
	TotalOpenMD      float64      `json:"total_open_md"`
	TeamFreeMD       float64      `json:"team_free_md"`
	GapMD            float64      `json:"gap_md"`
	NeedMorePeople   bool         `json:"need_more_people"`
	SuggestHeadcount int          `json:"suggest_headcount"`
	Members          []MemberLoad `json:"members"`
}

// MemberLoad is one member's workload for the manpower view: assigned manday on
// open tasks, remaining capacity in the current deadline window, and whether
// they are over their safe limit.
type MemberLoad struct {
	AccountID  string  `json:"account_id"`
	Name       string  `json:"name"`
	Position   string  `json:"position"`
	AssignedMD float64 `json:"assigned_md"`
	FreeMD     float64 `json:"free_md"`
	Over       bool    `json:"over"`
}

// SetMemberSkillsRequest is the body for an owner setting a member's skills.
type SetMemberSkillsRequest struct {
	Skills []profileEntity.ProfileSkill `json:"skills"`
}

// MemberSkillInfo is one member's current skills (for the PM editor).
type MemberSkillInfo struct {
	AccountID         string                       `json:"account_id"`
	Username          string                       `json:"username"`
	Name              string                       `json:"name"`
	Email             string                       `json:"email"`
	Role              string                       `json:"role"`
	Position          string                       `json:"position"`
	CapacityMDPerYear float64                      `json:"capacity_md_per_year"`
	Skills            []profileEntity.ProfileSkill `json:"skills"`
}

// MemberSkillsResponse lists every member of a project with their skills.
type MemberSkillsResponse struct {
	Members []MemberSkillInfo `json:"members"`
}

// SetMemberCapacityRequest is the body for PM editing a member's yearly MD.
type SetMemberCapacityRequest struct {
	CapacityMDPerYear float64 `json:"capacity_md_per_year" validate:"required,min=0,max=10000"`
}

// SetMemberRoleRequest is the body for changing a member's role (owner only).
type SetMemberRoleRequest struct {
	Role string `json:"role"` // "pm" or "member"
}
