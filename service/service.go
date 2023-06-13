package service

import (
	"github.com/google/go-github/v39/github"
	"github.com/gorilla/websocket"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
)

type Service struct {
	DB           *gorm.DB
	GitHubClient *github.Client
	GitLabClient *gitlab.Client

	WsClients map[*websocket.Conn]bool
}
