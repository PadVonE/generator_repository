package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
)

func (s *Service) CreateProject(ctx *gin.Context) {
	organizationId := ctx.Query("organization_id")
	projectName := ctx.Query("project_name")

	//	Поиск органицации в базе
	organization := entity.Organization{}
	err := s.DB.Model(&organization).Where("id = ?", organizationId).Take(&organization).Error

	_, err = s.getOrganizationRepositories(organization.Name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	repo, err := s.getRepository(organization.Name, projectName)

	project := s.createProject(repo, &organization)

	s.cloneProtoAndRealisation(project, organization)

}
