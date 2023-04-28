package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	"time"
)

type ContentItem struct {
	ID           int
	Organization string
	URL          string
	Created_at   time.Time
}

func (s *Service) Index(ctx *gin.Context) {

	organization := []entity.Organization{}

	err := s.DB.Model(&entity.Organization{}).Find(&organization).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		ctx.HTML(200, "new_organization", gin.H{})
		ctx.Next()
	}

	ctx.HTML(200, "index", gin.H{
		"Organization": organization,
	})
	ctx.Next()
}
