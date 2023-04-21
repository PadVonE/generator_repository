package service

import (
	"fmt"
	"generator/entity"
	"generator/usecase"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"os"
	"strings"
)

func (s *Service) CreateOrganisationApi(ctx *gin.Context) {

	orgName := "2q4t-plutus"

	//	Поиск органицации в базе
	organization := entity.Organization{}
	err := s.DB.Model(&organization).Where("name = ?", orgName).Take(&organization).Error
	organization.LocalPath = os.Getenv("GOPATH") + "/src/" + orgName

	if organization.Id == 0 {
		organization.Name = orgName
		err = s.DB.Create(&organization).Error

		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	}

	err = s.DB.Model(&organization).Where("name = ?", organization.Name).Take(&organization).Error

	repos, err := s.getOrganizationRepositories(organization.Name)
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
				commit, _ := s.getLastCommit(organization.Name, *repo.Name)
				release, _ := s.getLastRelease(organization.Name, *repo.Name)

				project := entity.Project{
					Type:             entity.GetTypeProjectByName(*repo.Name),
					OrganizationId:   organization.Id,
					Name:             repo.GetName(),
					LocalPath:        organization.LocalPath + "/implementation/" + strings.TrimPrefix(repo.GetName(), "proto-"),
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

func (s *Service) CreateOrganisationStructApi(ctx *gin.Context) {
	//
	organizationId := ctx.Query("organization_id")

	organization := entity.Organization{}

	err := s.DB.First(&organization, "id = ?", organizationId).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Создатём структуру папок для организации
	pathList := []string{
		"proto/github.com/" + organization.Name,
		"specification",
		"implementation",
		"docker",
	}

	for _, path := range pathList {

		p := organization.LocalPath + "/" + path

		err = os.MkdirAll(p, os.ModePerm)
		if err != nil {
			fmt.Printf("Ошибка при создании папок: %v\n", err)
			return
		}
	}
	// клонирование репозиториев proto в папку proto

	projects := []entity.Project{}
	query := s.DB.Model(&entity.Project{})
	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error
	for _, project := range projects {
		go func(project entity.Project, organization entity.Organization) {
			clonePath := ""
			switch project.Type {
			case entity.PROJECT_TYPE_REPOSITORY:
				clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + project.Name
			case entity.PROJECT_TYPE_USECASE:
				clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + project.Name

			case entity.PROJECT_TYPE_SPECIFICATION:
				clonePath = organization.LocalPath + "/specification/" + project.Name

			case entity.PROJECT_TYPE_NO_SET:
			}

			if clonePath != "" {
				err = usecase.CloningRepository(project.GithubUrl,
					clonePath,
					&http.BasicAuth{
						Username: envopt.GetEnv("GITHUB_USER"),
						Password: envopt.GetEnv("GITHUB_TOKEN"),
					})

				if err != nil {
					fmt.Printf("Error: %v\n", err)
				}
			}
		}(project, organization)
	}

}
