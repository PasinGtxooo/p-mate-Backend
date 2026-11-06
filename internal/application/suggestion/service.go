package suggestion

import (
	"context"
	"sort"
	"strings"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles"
	profileEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
	taskEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

// deadlineWindowMD is the assumed working capacity (in manday) one person has
// within the current sprint window. Kept as a constant for the hackathon; a
// production build would derive it from each member's calendar.
const deadlineWindowMD = 8.0

// requiredSkillLevel is the proficiency a candidate needs to be considered a
// full skill match for a task.
const requiredSkillLevel = 4

// defaultYearlyCapacityMD is the fallback capacity (in man-days per year) used
// when a profile hasn't set capacity_md_per_year. 240 ≈ 20 MD/month × 12 months,
// a typical consultant utilisation target.
const defaultYearlyCapacityMD = 240.0

// SuggestionService composes task, member and profile data to answer two
// questions for a PM: who should take a task, and does the team need more people.
type SuggestionService struct {
	taskRepo    tasks.TaskRepository
	projectRepo projects.ProjectRepository
	memberRepo  projects.ProjectMemberRepository
	profileRepo profiles.ProfileRepository
	accountRepo accounts.AccountRepository
}

func NewSuggestionService(
	taskRepo tasks.TaskRepository,
	projectRepo projects.ProjectRepository,
	memberRepo projects.ProjectMemberRepository,
	profileRepo profiles.ProfileRepository,
	accountRepo accounts.AccountRepository,
) *SuggestionService {
	return &SuggestionService{
		taskRepo:    taskRepo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
		profileRepo: profileRepo,
		accountRepo: accountRepo,
	}
}

// openStatuses are task statuses that still consume capacity.
func isOpen(status string) bool {
	return status == "todo" || status == "in_progress" || status == "in_review"
}

func mandayOf(t *taskEntity.Task) float64 {
	if t.EstimateManday != nil {
		return *t.EstimateManday
	}
	return 0
}

// skillLevelFor returns how well a profile matches a skill name (0 if absent).
func skillLevelFor(p *profileEntity.Profile, skill string) int {
	if skill == "" {
		return 0
	}
	for _, s := range p.Skills {
		if strings.EqualFold(s.Name, skill) {
			return s.Level
		}
	}
	return 0
}

// loadContext fetches the members, their profiles and all project tasks once,
// and verifies the caller is a member.
func (s *SuggestionService) loadContext(ctx context.Context, projectID uuid.UUID, nodeID string, accountID uuid.UUID) (
	[]*profileEntity.Profile, map[uuid.UUID]float64, []*taskEntity.Task, error) {

	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, nil, nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, nil, nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, nil, nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	members, err := s.projectRepo.ListProjectMembers(ctx, projectID, nodeUUID)
	if err != nil {
		return nil, nil, nil, apperror.NewInternalServerError("failed to list members", "LIST_MEMBERS_ERROR", err)
	}

	accountIDs := make([]string, 0, len(members))
	for _, m := range members {
		accountIDs = append(accountIDs, m.AccountID.String())
	}

	profs, err := s.profileRepo.ListProfilesByAccountIDs(ctx, accountIDs, nodeID)
	if err != nil {
		return nil, nil, nil, apperror.NewInternalServerError("failed to load profiles", "LOAD_PROFILES_ERROR", err)
	}

	projectTasks, err := s.taskRepo.ListTasksByProject(ctx, projectID)
	if err != nil {
		return nil, nil, nil, apperror.NewInternalServerError("failed to list tasks", "LIST_TASKS_ERROR", err)
	}

	// Sum assigned open manday per account. A task is "assigned" to a member
	// when its AssigneeID matches; tasks without an assignee are unassigned work.
	assigned := make(map[uuid.UUID]float64)
	for _, t := range projectTasks {
		if t.AssigneeID != nil && isOpen(t.Status) {
			assigned[*t.AssigneeID] += mandayOf(t)
		}
	}

	return profs, assigned, projectTasks, nil
}

// SuggestTeammatesForTask ranks members for a single task by skill then free capacity.
func (s *SuggestionService) SuggestTeammatesForTask(ctx context.Context, taskID uuid.UUID, nodeID string, accountID uuid.UUID) (*SuggestTeammateResponse, error) {
	tsk, err := s.taskRepo.GetTaskByID(ctx, taskID)
	if err != nil {
		return nil, apperror.NewNotFoundError("task not found", "TASK_NOT_FOUND", err)
	}

	profs, assigned, _, err := s.loadContext(ctx, tsk.ProjectID, nodeID, accountID)
	if err != nil {
		return nil, err
	}

	requiredSkill := ""
	if tsk.RequiredSkill != nil {
		requiredSkill = *tsk.RequiredSkill
	}
	taskMD := mandayOf(tsk)

	candidates := make([]TeammateCandidate, 0, len(profs))
	for _, p := range profs {
		lvl := skillLevelFor(p, requiredSkill)
		match := 0
		if requiredSkill != "" {
			match = lvl * 100 / 5
		}
		assignedMD := assigned[p.AccountID]
		freeMD := deadlineWindowMD - assignedMD
		fits := freeMD >= taskMD
		skillOK := requiredSkill == "" || lvl >= requiredSkillLevel

		candidates = append(candidates, TeammateCandidate{
			AccountID:   utils.ShortUUIDWithPrefix(p.AccountID, accountEntity.AccountIDPrefix),
			Name:        strings.TrimSpace(p.FirstName + " " + p.LastName),
			SkillLevel:  lvl,
			SkillMatch:  match,
			AssignedMD:  assignedMD,
			FreeMD:      freeMD,
			Fits:        fits,
			Recommended: skillOK && fits,
		})
	}

	// Rank: recommended first, then higher skill match, then more free capacity.
	sort.SliceStable(candidates, func(i, j int) bool {
		a, b := candidates[i], candidates[j]
		if a.Recommended != b.Recommended {
			return a.Recommended
		}
		if a.SkillMatch != b.SkillMatch {
			return a.SkillMatch > b.SkillMatch
		}
		return a.FreeMD > b.FreeMD
	})

	return &SuggestTeammateResponse{
		TaskID:         utils.ShortUUIDWithPrefix(tsk.ID, taskEntity.TaskIDPrefix),
		TaskName:       tsk.Name,
		RequiredSkill:  requiredSkill,
		EstimateManday: taskMD,
		Candidates:     candidates,
	}, nil
}

// GetManpower reports whether the team can absorb all open manday before the deadline.
func (s *SuggestionService) GetManpower(ctx context.Context, projectID uuid.UUID, nodeID string, accountID uuid.UUID) (*ManpowerResponse, error) {
	profs, assigned, projectTasks, err := s.loadContext(ctx, projectID, nodeID, accountID)
	if err != nil {
		return nil, err
	}

	var totalOpenMD float64
	for _, t := range projectTasks {
		if isOpen(t.Status) {
			totalOpenMD += mandayOf(t)
		}
	}

	var teamFreeMD float64
	for _, p := range profs {
		free := deadlineWindowMD - assigned[p.AccountID]
		if free > 0 {
			teamFreeMD += free
		}
	}

	gap := totalOpenMD - teamFreeMD
	need := gap > 0
	headcount := 0
	if need {
		headcount = int((gap + deadlineWindowMD - 1) / deadlineWindowMD) // ceil
	}

	members := make([]MemberLoad, 0, len(profs))
	for _, p := range profs {
		md := assigned[p.AccountID]
		pos := ""
		if p.Position != nil {
			pos = *p.Position
		}
		members = append(members, MemberLoad{
			AccountID:  utils.ShortUUIDWithPrefix(p.AccountID, accountEntity.AccountIDPrefix),
			Name:       strings.TrimSpace(p.FirstName + " " + p.LastName),
			Position:   pos,
			AssignedMD: md,
			FreeMD:     deadlineWindowMD - md,
			Over:       md > deadlineWindowMD,
		})
	}

	return &ManpowerResponse{
		ProjectID:        utils.ShortUUIDWithPrefix(projectID, "proj"),
		TotalOpenMD:      totalOpenMD,
		TeamFreeMD:       teamFreeMD,
		GapMD:            maxF(gap, 0),
		NeedMorePeople:   need,
		SuggestHeadcount: headcount,
		Members:          members,
	}, nil
}

func maxF(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// SetMemberSkills lets a project owner set another member's skills.
// Only the owner of the project may do this; a member may still edit their own
// skills through the normal profile endpoint.
func (s *SuggestionService) SetMemberSkills(ctx context.Context, projectID uuid.UUID, targetAccountID uuid.UUID, nodeID string, callerAccountID uuid.UUID, skills []profileEntity.ProfileSkill) error {
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Caller must be the project owner.
	caller, err := s.projectRepo.GetProjectMember(ctx, projectID, callerAccountID, nodeUUID)
	if err != nil {
		return apperror.NewForbiddenError("you don't have access to this project", "NO_ACCESS", err)
	}
	if !caller.Role.CanManageProject() {
		return apperror.NewForbiddenError("only the project owner or PM can edit member skills", "NOT_MANAGER", nil)
	}

	// Target must be a member of the project.
	if _, err := s.projectRepo.GetProjectMember(ctx, projectID, targetAccountID, nodeUUID); err != nil {
		return apperror.NewNotFoundError("member not found in project", "MEMBER_NOT_FOUND", err)
	}

	if skills == nil {
		skills = []profileEntity.ProfileSkill{}
	}
	if err := s.profileRepo.UpdateProfileSkills(ctx, targetAccountID.String(), nodeID, skills); err != nil {
		return apperror.NewInternalServerError("failed to update member skills", "UPDATE_SKILLS_ERROR", err)
	}
	return nil
}

// ListMemberSkills returns every project member with their current skills,
// for the owner's skill editor. Any member of the project may read it.
func (s *SuggestionService) ListMemberSkills(ctx context.Context, projectID uuid.UUID, nodeID string, accountID uuid.UUID) (*MemberSkillsResponse, error) {
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	members, err := s.projectRepo.ListProjectMembers(ctx, projectID, nodeUUID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list members", "LIST_MEMBERS_ERROR", err)
	}

	accountIDs := make([]string, 0, len(members))
	roleByID := make(map[string]string, len(members))
	for _, m := range members {
		accountIDs = append(accountIDs, m.AccountID.String())
		roleByID[m.AccountID.String()] = m.Role.String()
	}

	profs, err := s.profileRepo.ListProfilesByAccountIDs(ctx, accountIDs, nodeID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to load profiles", "LOAD_PROFILES_ERROR", err)
	}
	profByID := make(map[string]*profileEntity.Profile, len(profs))
	for _, p := range profs {
		profByID[p.AccountID.String()] = p
	}

	// Look up username and email for every member so the UI can always show a readable
	// label, even when a member hasn't created a profile yet.
	usernameByID := make(map[string]string, len(members))
	emailByID := make(map[string]string, len(members))
	if s.accountRepo != nil {
		accs, err := s.accountRepo.ListByIDs(ctx, accountIDs)
		if err != nil {
			return nil, apperror.NewInternalServerError("failed to load accounts", "LOAD_ACCOUNTS_ERROR", err)
		}
		for _, a := range accs {
			usernameByID[a.ID.String()] = a.Username
			emailByID[a.ID.String()] = a.Email
		}
	}

	items := make([]MemberSkillInfo, 0, len(members))
	for _, m := range members {
		idStr := m.AccountID.String()
		info := MemberSkillInfo{
			AccountID: utils.ShortUUIDWithPrefix(m.AccountID, accountEntity.AccountIDPrefix),
			Username:  usernameByID[idStr],
			Email:     emailByID[idStr],
			Role:      roleByID[idStr],
			Skills:    []profileEntity.ProfileSkill{},
		}
		if p := profByID[idStr]; p != nil {
			info.Name = strings.TrimSpace(p.FirstName + " " + p.LastName)
			if p.Position != nil {
				info.Position = *p.Position
			}
			if p.CapacityMDPerYear != nil {
				info.CapacityMDPerYear = *p.CapacityMDPerYear
			} else {
				info.CapacityMDPerYear = defaultYearlyCapacityMD
			}
			if p.Skills != nil {
				info.Skills = p.Skills
			}
		}
		items = append(items, info)
	}

	return &MemberSkillsResponse{Members: items}, nil
}

// SetMemberCapacity lets the owner or a PM set a member's yearly capacity
// (in man-days). Used by the ranker to decide who is free in the task's window.
func (s *SuggestionService) SetMemberCapacity(ctx context.Context, projectID uuid.UUID, targetAccountID uuid.UUID, nodeID string, callerAccountID uuid.UUID, capacityMDPerYear float64) error {
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}
	if capacityMDPerYear < 0 || capacityMDPerYear > 10000 {
		return apperror.NewBadRequestError("capacity must be between 0 and 10000 MD/year", "INVALID_CAPACITY", nil)
	}

	caller, err := s.projectRepo.GetProjectMember(ctx, projectID, callerAccountID, nodeUUID)
	if err != nil {
		return apperror.NewForbiddenError("you don't have access to this project", "NO_ACCESS", err)
	}
	if caller.Role != projectEntity.RoleOwner && caller.Role != projectEntity.RolePM {
		return apperror.NewForbiddenError("only owner or PM can set capacity", "NOT_MANAGER", nil)
	}

	if err := s.profileRepo.UpdateCapacity(ctx, targetAccountID.String(), nodeID, capacityMDPerYear); err != nil {
		return apperror.NewInternalServerError("failed to update capacity", "UPDATE_CAPACITY_ERROR", err)
	}
	return nil
}

// SetMemberRole lets the project owner promote a member to PM or demote back to
// member. Only the owner may change roles, and the owner's own role is fixed.
func (s *SuggestionService) SetMemberRole(ctx context.Context, projectID uuid.UUID, targetAccountID uuid.UUID, nodeID string, callerAccountID uuid.UUID, role string) error {
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	newRole := projectEntity.ProjectRole(role)
	if newRole != projectEntity.RolePM && newRole != projectEntity.RoleMember {
		return apperror.NewBadRequestError("role must be 'pm' or 'member'", "INVALID_ROLE", nil)
	}

	caller, err := s.projectRepo.GetProjectMember(ctx, projectID, callerAccountID, nodeUUID)
	if err != nil {
		return apperror.NewForbiddenError("you don't have access to this project", "NO_ACCESS", err)
	}
	if caller.Role != projectEntity.RoleOwner {
		return apperror.NewForbiddenError("only the project owner can change member roles", "NOT_OWNER", nil)
	}

	target, err := s.projectRepo.GetProjectMember(ctx, projectID, targetAccountID, nodeUUID)
	if err != nil {
		return apperror.NewNotFoundError("member not found in project", "MEMBER_NOT_FOUND", err)
	}
	if target.Role == projectEntity.RoleOwner {
		return apperror.NewBadRequestError("cannot change the owner's role", "CANNOT_CHANGE_OWNER", nil)
	}

	if err := s.projectRepo.UpdateMemberRole(ctx, projectID, targetAccountID, nodeUUID, newRole); err != nil {
		return apperror.NewInternalServerError("failed to update member role", "UPDATE_ROLE_ERROR", err)
	}
	return nil
}
