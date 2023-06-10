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
	Id                     int32
	CreatedAt              time.Time `gorm:"->;<-:create"`
	UpdatedAt              time.Time
	PushedAt               time.Time
	Type                   int32
	OrganizationId         int32
	Name                   string
	LocalPath              string
	GithubUrl              string
	GithubLastCommitName   string
	GithubLastCommitTime   time.Time
	GithubLastCommitAuthor string
	GithubReleaseTag       string

	GitlabUrl              string
	GitlabLastCommitName   string
	GitlabLastCommitTime   time.Time
	GitlabLastCommitAuthor string
	GitlabReleaseTag       string

	LastStructure string

	LastStructureUnmarshal ProjectComponents `gorm:"-"`
	NewTag                 string            `gorm:"-"`
	NewCommitName          string            `gorm:"-"`
	NewCommitDate          time.Time         `gorm:"-"`
	HasClone               bool              `gorm:"-"`
	IsNewProject           bool              `gorm:"-"`
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

func GetPath(projectType int32, projectName string, organization *Organization) (clonePath, repositoryRealisation string) {
	switch projectType {
	case PROJECT_TYPE_REPOSITORY:
		clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + projectName
		repositoryRealisation = strings.TrimPrefix(projectName, "proto-")
	case PROJECT_TYPE_USECASE:
		clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + projectName
		repositoryRealisation = strings.TrimPrefix(projectName, "proto-")

	case PROJECT_TYPE_SPECIFICATION:
		clonePath = organization.LocalPath + "/specification/" + projectName
		repositoryRealisation = "gateway-" + strings.TrimPrefix(projectName, "specification-")

	case PROJECT_TYPE_NO_SET:
	}
	return
}
