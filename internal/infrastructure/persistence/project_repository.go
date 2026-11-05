package persistence

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) projects.ProjectRepository {
	return &projectRepository{db: db}
}

func (r *projectRepository) CreateProject(ctx context.Context, proj *entity.Project) error {
	return r.db.WithContext(ctx).Create(proj).Error
}

func (r *projectRepository) GetProjectByID(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) (*entity.Project, error) {
	var proj entity.Project
	err := r.db.WithContext(ctx).
		Where("id = ?", projectID).
		First(&proj).Error
	if err != nil {
		return nil, err
	}
	return &proj, nil
}

func (r *projectRepository) ListProjectByAccountID(ctx context.Context, accountID uuid.UUID, nodeID uuid.UUID, limit, offset int) ([]*entity.Project, int, error) {
	var projects []*entity.Project
	var total int64

	baseQuery := r.db.WithContext(ctx).
		Model(&entity.Project{}).
		Where("projects.owner_id = ? OR projects.id IN (?)",
			accountID,
			r.db.Table("project_members").
				Select("project_id").
				Where("account_id = ?", accountID),
		)

	// Get total count
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("projects.owner_id = ? OR projects.id IN (?)",
			accountID,
			r.db.Table("project_members").
				Select("project_id").
				Where("account_id = ?", accountID),
		).
		Limit(limit).
		Offset(offset).
		Find(&projects).Error

	return projects, int(total), err
}

func (r *projectRepository) UpdateProject(ctx context.Context, proj *entity.Project, nodeID uuid.UUID) error {
	// Verify the project belongs to the tenant before updating
	return r.db.WithContext(ctx).
		Model(&entity.Project{}).
		Select("*").
		Where("id = ? AND node_id = ?", proj.ID, nodeID).
		Updates(proj).Error
}

func (r *projectRepository) DeleteProject(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) error {
	// Verify the project belongs to the tenant before deleting
	return r.db.WithContext(ctx).
		Where("id = ? AND node_id = ?", projectID, nodeID).
		Delete(&entity.Project{}).Error
}

func (r *projectRepository) AddMember(ctx context.Context, member *entity.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

func (r *projectRepository) RemoveMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND account_id = ? AND node_id = ? AND role != ?", projectID, accountID, nodeID, entity.RoleOwner).
		Delete(&entity.ProjectMember{}).Error
}

func (r *projectRepository) GetProjectMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID) (*entity.ProjectMember, error) {
	var member entity.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ? AND account_id = ? AND node_id = ?", projectID, accountID, nodeID).
		First(&member).Error
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func (r *projectRepository) ListProjectMembers(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) ([]*entity.ProjectMember, error) {
	var members []*entity.ProjectMember
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&members).Error
	return members, err
}

func (r *projectRepository) UpdateMemberRole(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID, role entity.ProjectRole) error {
	return r.db.WithContext(ctx).
		Model(&entity.ProjectMember{}).
		Where("project_id = ? AND account_id = ? AND node_id = ?", projectID, accountID, nodeID).
		Update("role", role).Error
}
