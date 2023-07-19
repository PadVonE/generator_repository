package service

import (
	"fmt"
	"generator/entity"
	"generator/generators/docker"
	"github.com/gin-gonic/gin"
	"strconv"
	"strings"
)

func (s *Service) GenerateDockerApi(ctx *gin.Context) {
	organizationId := ctx.Query("organization_id")

	// Получаем список проектов
	projects := []entity.Project{}

	query := s.DB.Model(&entity.Project{})

	err := query.Where("organization_id = ?", organizationId).Order("id ASC").Find(&projects).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	dockerCompose := "version: '3'\nservices:\n\n  #DB-------------------------------------------------------------"

	for _, project := range projects {
		if project.Type == entity.PROJECT_TYPE_REPOSITORY {

			name := strings.ReplaceAll(project.Name, "proto-", "")
			port := strconv.Itoa(16000 + int(project.Id))

			code, err := generators.GenerateDockerComposeDatabase(name, port)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			dockerCompose = dockerCompose + " \n\n " + code
		}

	}

	ctx.JSON(200, gin.H{
		"code": dockerCompose,
		"err":  err,
	})

}
