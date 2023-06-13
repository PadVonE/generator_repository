package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"path/filepath"
)

func (s *Service) GeneratePathProjectRepositoryApi(ctx *gin.Context) {

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	servicePath := filepath.FromSlash(project.LocalPath)
	pathList := []string{
		"entity",
		"migrations",
		"service",
		"healthcheck",
	}

	err = os.MkdirAll(servicePath, os.ModePerm)
	if err != nil {
		log.Error("Ошибка при создании папок: %v\n", err)
		return
	}

	for _, path := range pathList {

		p := servicePath + "/" + path

		err = os.MkdirAll(p, os.ModePerm)
		if err != nil {
			log.Error("Ошибка при создании папок: %v\n", err)
			return
		}
	}
}
