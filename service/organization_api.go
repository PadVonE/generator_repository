package service

import (
	"fmt"
	"generator/entity"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
)

func (s *Service) CreateOrganisationApi(ctx *gin.Context) {

	orgName := "2q4t-plutus"

	//	Поиск органицации в базе
	organization := entity.Organization{}
	err := s.DB.Model(&organization).Where("name = ?", orgName).Take(&organization).Error

	if organization.Id == 0 {
		organization.Name = orgName
		err = s.DB.Create(&organization).Error

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	err = s.DB.Model(&organization).Where("name = ?", organization.Name).Take(&organization).Error

	repos, err := s.getOrganizationRepositories(organization.Name, envopt.GetEnv("GITHUB_TOKEN"))
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	projects := []entity.Project{}

	query := s.DB.Model(&entity.Project{})

	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, repo := range repos {
		go func(repo *github.Repository, organization entity.Organization, projects []entity.Project) {
			hasDb := false
			for _, project := range projects {
				if project.Name == *repo.Name {
					hasDb = true
				}
			}

			if !hasDb {
				commit, _ := s.getLastCommit(organization.Name, *repo.Name, envopt.GetEnv("GITHUB_TOKEN"))
				release, _ := s.getLastRelease(organization.Name, *repo.Name, envopt.GetEnv("GITHUB_TOKEN"))

				project := entity.Project{
					Type:             entity.GetTypeProjectByName(*repo.Name),
					OrganizationId:   organization.Id,
					Name:             repo.GetName(),
					DirPath:          repo.GetName(),
					GithubUrl:        "https://github.com/" + organization.Name + "/" + repo.GetName(),
					PushedAt:         repo.GetPushedAt().UTC(),
					LastCommitName:   commit.Commit.GetMessage(),
					LastCommitTime:   commit.Commit.GetAuthor().GetDate(),
					LastCommitAuthor: commit.Commit.GetAuthor().GetName(),
					ReleaseTag:       release.GetTagName(),
					LastStructure:    "{}",
				}

				s.DB.Create(&project)
			}
		}(repo, organization, projects)
	}

	ctx.Next()
}
