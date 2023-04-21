package entity

import (
	"strings"
	"time"
)

type Project struct {
	Id               int32
	CreatedAt        time.Time `gorm:"->;<-:create"`
	UpdatedAt        time.Time
	PushedAt         time.Time
	Type             int32
	OrganizationId   int32
	Name             string
	DirPath          string
	GithubUrl        string
	LastCommitName   string
	LastCommitTime   time.Time
	LastCommitAuthor string
	ReleaseTag       string
	LastStructure    string

	NewTag string `gorm:"-"`
}

func (project *Project) TableName() string {
	return "project"
}

func GetTypeProjectByName(name string) int32 {

	if strings.HasSuffix(name, "repository") {
		return 1
	}

	if strings.HasSuffix(name, "usecase") {
		return 2
	}

	if strings.HasPrefix(name, "specification") {
		return 3
	}

	return 0

}
