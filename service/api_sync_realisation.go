package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	"github.com/xanzy/go-gitlab"
)

func (s *Service) SyncRealisationApi(ctx *gin.Context) {
	projectID := ctx.Query("project_id")

	project := entity.Project{}

	query := s.DB.Model(&project)

	err := query.Where("id = ?", projectID).Take(&project).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	organization := entity.Organization{}

	query = s.DB.Model(&organization)

	err = query.Where("id = ?", project.OrganizationId).Take(&organization).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	repositoryRealisation := entity.GetRealisationName(project.Type, project.Name)
	if repositoryRealisation == "" {
		repositoryRealisation = project.Name
	}

	// Проверяем есть ли реализация
	exist := true
	realisationRepo, _, err := s.GitLabClient.Projects.GetProject(organization.GitlabOrganizationName()+"/"+repositoryRealisation, nil)
	if err != nil {
		if errResponse, ok := err.(*gitlab.ErrorResponse); ok && errResponse.Response.StatusCode == 404 {
			exist = false
		}
		exist = false
	}

	if exist {
		release, _ := s.getLastReleaseGitlab(realisationRepo)

		if release != nil && release.Name !=
			project.RealisationReleaseTag {
			project.RealisationReleaseTag = release.Name
		}

		commit, _ := s.getLastCommitGitlab(realisationRepo, "dev")

		project.RealisationLastCommitName = commit.Title
		project.RealisationLastCommitAuthor = commit.AuthorName
		project.RealisationLastCommitTime = *realisationRepo.LastActivityAt
		project.RealisationUrl = organization.GitlabUrl + repositoryRealisation
	}
	err = s.DB.Save(&project).Error

	ctx.JSON(200, gin.H{
		"struct": project,
		"err":    err,
	})

}
