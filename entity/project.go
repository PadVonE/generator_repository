package entity

import (
	"generator/helpers"
	"strings"
	"time"
)

const PROJECT_TYPE_NO_SET = 0
const PROJECT_TYPE_REPOSITORY = 1
const PROJECT_TYPE_USECASE = 2
const PROJECT_TYPE_SPECIFICATION = 3

type Project struct {
	Id             int32
	CreatedAt      time.Time `gorm:"->;<-:create"`
	UpdatedAt      time.Time
	PushedAt       time.Time
	Type           int32
	OrganizationId int32
	Name           string
	LocalPath      string

	SpecificationUrl              string
	SpecificationLastCommitName   string
	SpecificationLastCommitTime   time.Time
	SpecificationLastCommitAuthor string
	SpecificationReleaseTag       string

	RealisationUrl              string
	RealisationLastCommitName   string
	RealisationLastCommitTime   time.Time
	RealisationLastCommitAuthor string
	RealisationReleaseTag       string

	LastStructure string

	LastStructureUnmarshal       ProjectComponents `gorm:"-"`
	NewSpecificationTag          string            `gorm:"-"`
	NewSpecificationCommitAuthor string            `gorm:"-"`
	NewSpecificationCommitName   string            `gorm:"-"`
	NewSpecificationCommitDate   time.Time         `gorm:"-"`

	SpecificationRepoInfo *helpers.GitRepoInfo `gorm:"-"`

	NewRealisationTag          string    `gorm:"-"`
	NewRealisationCommitAuthor string    `gorm:"-"`
	NewRealisationCommitName   string    `gorm:"-"`
	NewRealisationCommitDate   time.Time `gorm:"-"`

	RealisationRepoInfo *helpers.GitRepoInfo `gorm:"-"`

	HasClone     bool `gorm:"-"`
	IsNewProject bool `gorm:"-"`
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

func GetPath(projectType int32, projectName string, organization *Organization) (clonePath string) {
	switch projectType {
	case PROJECT_TYPE_REPOSITORY:
		clonePath = organization.LocalPath + "/proto/" + projectName
	case PROJECT_TYPE_USECASE:
		clonePath = organization.LocalPath + "/proto/" + projectName
	case PROJECT_TYPE_SPECIFICATION:
		clonePath = organization.LocalPath + "/specification/" + projectName

	case PROJECT_TYPE_NO_SET:
	}
	return
}

func GetRealisationName(projectType int32, projectName string) (repositoryRealisation string) {
	switch projectType {
	case PROJECT_TYPE_REPOSITORY:
		repositoryRealisation = strings.TrimPrefix(projectName, "proto-")
	case PROJECT_TYPE_USECASE:
		repositoryRealisation = strings.TrimPrefix(projectName, "proto-")
	case PROJECT_TYPE_SPECIFICATION:
		repositoryRealisation = "gateway-" + strings.TrimPrefix(projectName, "specification-")

	case PROJECT_TYPE_NO_SET:
	}
	return
}

func (project *Project) GetRealisationName() (repositoryRealisation string) {
	switch project.Type {
	case PROJECT_TYPE_REPOSITORY:
		repositoryRealisation = strings.TrimPrefix(project.Name, "proto-")
	case PROJECT_TYPE_USECASE:
		repositoryRealisation = strings.TrimPrefix(project.Name, "proto-")
	case PROJECT_TYPE_SPECIFICATION:
		repositoryRealisation = "gateway-" + strings.TrimPrefix(project.Name, "specification-")

	case PROJECT_TYPE_NO_SET:
		repositoryRealisation = project.Name
	}
	return
}
