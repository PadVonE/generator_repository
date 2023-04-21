package entity

import (
	"time"
)

type CloneHistory struct {
	Id          int32
	CreatedAt   time.Time `gorm:"->;<-:create"`
	ProjectId   int32
	Name        string
	CloningPath string
	ReleaseTag  string
	Structure   string
}

func (project *CloneHistory) TableName() string {
	return "clone_history"
}
