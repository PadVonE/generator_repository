package service

import (
	"fmt"
	"generator/entity"
	"generator/helpers"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func (s *Service) GitCheckout(ctx *gin.Context) {
	projectID := ctx.Query("project_id")
	branch := ctx.Query("branch")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	err = helpers.GitCheckoutBranch(project.LocalPath, branch)
	if err != nil {
		log.Error(err)
	}

	ctx.JSON(200, gin.H{
		"err": err,
	})

}

func (s *Service) GitCreateBranch(ctx *gin.Context) {
	projectID := ctx.Query("project_id")
	branch := ctx.Query("branch")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	err = helpers.CreateBranch(project.LocalPath, branch)
	if err != nil {
		log.Error(err)
	}

	ctx.JSON(200, gin.H{
		"err": err,
	})

}
