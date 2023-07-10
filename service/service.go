package service

import (
	"github.com/google/go-github/v39/github"
	"github.com/gorilla/websocket"
	"github.com/xanzy/go-gitlab"
	"gorm.io/gorm"
	"sync"
)

type Service struct {
	DB           *gorm.DB
	GitHubClient *github.Client
	GitLabClient *gitlab.Client

	WsClients map[*WsClient]bool
}

type WsClient struct {
	Conn  *websocket.Conn
	Mutex sync.Mutex
}
