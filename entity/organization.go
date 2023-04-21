package entity

import (
	"time"
)

type Organization struct {
	Id         int32
	CreatedAt  time.Time `gorm:"->;<-:create"`
	UpdatedAt  time.Time
	LastUpdate time.Time
	Name       string
}

func (card *Organization) TableName() string {
	return "organization"
}
