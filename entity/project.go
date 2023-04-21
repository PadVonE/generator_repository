package entity

import (
	"strings"
	"time"
)

const PROJECT_TYPE_NO_SET = 0
const PROJECT_TYPE_REPOSITORY = 1
const PROJECT_TYPE_USECASE = 2
const PROJECT_TYPE_SPECIFICATION = 3

type Project struct {
	Id               int32
	CreatedAt        time.Time `gorm:"->;<-:create"`
	UpdatedAt        time.Time
	PushedAt         time.Time
	Type             int32
	OrganizationId   int32
	Name             string
	LocalPath        string
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
		return PROJECT_TYPE_REPOSITORY
	}

	if strings.HasSuffix(name, "usecase") {
		return PROJECT_TYPE_USECASE
	}

	if strings.HasPrefix(name, "specification") {
		return PROJECT_TYPE_SPECIFICATION
	}

	return PROJECT_TYPE_NO_SET

}
