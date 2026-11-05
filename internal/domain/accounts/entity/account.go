package entity

import (
	"time"

	"github.com/google/uuid"
)

const AccountIDPrefix = "acc"

type Account struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	NodeID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Username  string    `gorm:"type:varchar(100);unique;not null"`
	Email     string    `gorm:"type:varchar(255);unique;not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	State     string    `gorm:"type:enum('active','inactive');not null;default:'active'"`
	CreatedAt time.Time `gorm:"not null"`
	UpdatedAt time.Time `gorm:"not null"`
}

func (Account) TableName() string {
	return "accounts"
}
