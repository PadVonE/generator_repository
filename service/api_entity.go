package service

import (
	"github.com/gin-gonic/gin"
)

func (s *Service) EntityGenerator(ctx *gin.Context) {

	ctx.JSON(200, []string{})
}
