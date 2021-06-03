package service

import (
	"generator/entity"
	"github.com/gin-gonic/gin"
)

func (s *Service) Index(ctx *gin.Context) {

	viewData := entity.ViewData{}

	// Common Data
	SetTemplate(ctx, "index")
	SetPayload(ctx, viewData)

	// Common Data
	ctx.Next()
}
