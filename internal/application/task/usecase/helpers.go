package usecase

import (
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/google/uuid"
)

// mapAssignee converts a stored assignee UUID into an acc_ prefixed string for API responses.
func mapAssignee(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := utils.ShortUUIDWithPrefix(*id, accountEntity.AccountIDPrefix)
	return &s
}
