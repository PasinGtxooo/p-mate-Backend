package entity

import (
	"database/sql/driver"
	"fmt"
)

// ProjectRole represents the role of a member in a project
type ProjectRole string

const (
	RoleOwner  ProjectRole = "owner"
	RolePM     ProjectRole = "pm"
	RoleMember ProjectRole = "member"
)

// Valid roles
var validRoles = map[ProjectRole]bool{
	RoleOwner:  true,
	RolePM:     true,
	RoleMember: true,
}

// IsValid checks if the role is valid
func (r ProjectRole) IsValid() bool {
	return validRoles[r]
}

// CanManageProject reports whether this role may manage the project's plan:
// allocate work, edit member skills, and so on. Owner and PM can; a plain
// member cannot.
func (r ProjectRole) CanManageProject() bool {
	return r == RoleOwner || r == RolePM
}

// String returns the string representation of the role
func (r ProjectRole) String() string {
	return string(r)
}

// Scan implements the sql.Scanner interface for database deserialization
func (r *ProjectRole) Scan(value interface{}) error {
	if value == nil {
		*r = RoleMember // default value
		return nil
	}

	str, ok := value.(string)
	if !ok {
		return fmt.Errorf("failed to scan ProjectRole: expected string, got %T", value)
	}

	role := ProjectRole(str)
	if !role.IsValid() {
		return fmt.Errorf("invalid role: %s", str)
	}

	*r = role
	return nil
}

// Value implements the driver.Valuer interface for database serialization
func (r ProjectRole) Value() (driver.Value, error) {
	if !r.IsValid() {
		return nil, fmt.Errorf("invalid role: %s", r)
	}
	return string(r), nil
}
