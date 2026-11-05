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
	ProjectID        string  `json:"project_id"`
	TotalOpenMD      float64 `json:"total_open_md"`
	TeamFreeMD       float64 `json:"team_free_md"`
	GapMD            float64 `json:"gap_md"`
	NeedMorePeople   bool    `json:"need_more_people"`
	SuggestHeadcount int     `json:"suggest_headcount"`
}

// SetMemberSkillsRequest is the body for an owner setting a member's skills.
type SetMemberSkillsRequest struct {
	Skills []profileEntity.ProfileSkill `json:"skills"`
}

// MemberSkillInfo is one member's current skills (for the PM editor).
type MemberSkillInfo struct {
	AccountID string                       `json:"account_id"`
	Username  string                       `json:"username"`
	Name      string                       `json:"name"`
	Role      string                       `json:"role"`
	Skills    []profileEntity.ProfileSkill `json:"skills"`
}

// MemberSkillsResponse lists every member of a project with their skills.
type MemberSkillsResponse struct {
	Members []MemberSkillInfo `json:"members"`
}

// SetMemberRoleRequest is the body for changing a member's role (owner only).
type SetMemberRoleRequest struct {
	Role string `json:"role"` // "pm" or "member"
}
