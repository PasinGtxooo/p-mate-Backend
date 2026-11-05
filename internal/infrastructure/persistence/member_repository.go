package persistence

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type memberRepository struct {
	db *gorm.DB
}

func NewMemberRepository(db *gorm.DB) projects.ProjectMemberRepository {
	return &memberRepository{db: db}
}

// Create represent creates a new project member
func (r *memberRepository) Create(ctx context.Context, member *entity.ProjectMember) error {
	return r.db.WithContext(ctx).Create(member).Error
}

// IsMember checks if an account is a member of a project
func (r *memberRepository) IsMember(ctx context.Context, projectID, accountID uuid.UUID) (bool, error) {
	var memberCount int64
	err := r.db.WithContext(ctx).
		Model(&entity.ProjectMember{}).
		Where("project_id = ? AND account_id = ? AND role IN (?)", projectID, accountID, []entity.ProjectRole{entity.RoleOwner, entity.RolePM, entity.RoleMember}).
		Count(&memberCount).Error

	if err != nil {
		return false, err
	}

	return memberCount > 0, nil
}
