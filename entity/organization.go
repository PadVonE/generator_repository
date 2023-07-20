package entity

import (
	"strings"
	"time"
)

type Organization struct {
	Id          int32
	CreatedAt   time.Time `gorm:"->;<-:create"`
	UpdatedAt   time.Time
	LastUpdate  time.Time
	Name        string
	GithubUrl   string
	GitlabUrl   string
	LocalPath   string
	JiraProject string
}

func (org *Organization) TableName() string {
	return "organization"
}

func (org *Organization) GithubOrganizationName() string {
	parts := strings.Split(org.GithubUrl, "/")
	name := parts[len(parts)-1]

	return name
}

func (org *Organization) GitlabOrganizationName() string {
	name := strings.ReplaceAll(org.GitlabUrl, "https://gitlab.com/", "")
	name = strings.TrimSuffix(name, "/")
	return name
}
