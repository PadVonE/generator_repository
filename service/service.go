package service

import (
	"github.com/google/go-github/v39/github"
	"gorm.io/gorm"
)

type Service struct {
	DB        *gorm.DB
	GitClient *github.Client
}
