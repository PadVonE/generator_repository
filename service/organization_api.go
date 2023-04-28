package service

import (
	"fmt"
	"generator/entity"
	"generator/helpers"
	"generator/usecase"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"os"
	"strings"
	"sync"
)

type CreateOrganizationRequest struct {
	Name      string `json:"name" binding:"required"`
	GithubUrl string `json:"github_url" binding:"required"`
	GitlabUrl string `json:"gitlab_url"`
	LocalPath string `json:"local_path" binding:"required"`
}

func (s *Service) ListOrganizationApi(ctx *gin.Context) {

	//	Поиск органицации в базе
	organization := entity.Organization{}
	err := s.DB.Model(&organization).Where("name = ?", organization.Name).Take(&organization).Error

	_, err = s.getOrganizationRepositories(organization.Name)
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

	ctx.Next()
}

func (s *Service) CreateOrganizationStructApi(ctx *gin.Context) {
	//
	var err error
	var req CreateOrganizationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": err.Error()})
		return
	}

	organization := entity.Organization{}
	err = s.DB.Model(&organization).Where("name = ?", req.Name).Take(&organization).Error

	if organization.Id == 0 {

		organization = entity.Organization{
			Name:      req.Name,
			GithubUrl: req.GithubUrl,
			GitlabUrl: req.GitlabUrl,
			LocalPath: req.LocalPath,
		}

		if err := s.DB.Create(&organization).Error; err != nil {
			ctx.JSON(500, gin.H{"error": err.Error()})
			return
		}
	}

	// Создаём прокеты в базе данных для организации

	repos, err := s.getOrganizationRepositories(organization.Name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	projects := []entity.Project{}

	query := s.DB.Model(&entity.Project{})

	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error

	var wg sync.WaitGroup

	for _, repo := range repos {
		wg.Add(1)
		go func(repo *github.Repository, organization *entity.Organization, projects []entity.Project) {
			defer wg.Done()
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
					Type:                   entity.GetTypeProjectByName(*repo.Name),
					OrganizationId:         organization.Id,
					Name:                   repo.GetName(),
					LocalPath:              organization.LocalPath + "/implementation/" + strings.TrimPrefix(repo.GetName(), "proto-"),
					GithubUrl:              "https://github.com/" + organization.Name + "/" + repo.GetName(),
					PushedAt:               repo.GetPushedAt().UTC(),
					GithubLastCommitName:   commit.Commit.GetMessage(),
					GithubLastCommitTime:   commit.Commit.GetAuthor().GetDate(),
					GithubLastCommitAuthor: commit.Commit.GetAuthor().GetName(),
					GithubReleaseTag:       release.GetTagName(),
					LastStructure:          "{}",
				}

				s.DB.Create(&project)
			}
		}(repo, &organization, projects)
	}
	wg.Wait()

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

	projects = []entity.Project{}

	query = s.DB.Model(&entity.Project{})
	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error
	for _, project := range projects {
		wg.Add(1)
		go func(project entity.Project, organization entity.Organization) {
			defer wg.Done()
			repositoryRealisation := ""

			clonePath := ""
			switch project.Type {
			case entity.PROJECT_TYPE_REPOSITORY:
				clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + project.Name
				repositoryRealisation = strings.TrimPrefix(project.Name, "proto-")
			case entity.PROJECT_TYPE_USECASE:
				clonePath = organization.LocalPath + "/proto/github.com/" + organization.Name + "/" + project.Name
				repositoryRealisation = strings.TrimPrefix(project.Name, "proto-")

			case entity.PROJECT_TYPE_SPECIFICATION:
				clonePath = organization.LocalPath + "/specification/" + project.Name
				repositoryRealisation = "gateway-" + strings.TrimPrefix(project.Name, "specification-")

			case entity.PROJECT_TYPE_NO_SET:
			}

			if len(clonePath) > 0 {
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

			if len(repositoryRealisation) > 0 {
				exist, err := helpers.RepoExists(envopt.GetEnv("GITLAB_TOKEN"), organization.GitlabOrganizationName()+"/"+repositoryRealisation)
				if exist {
					fmt.Printf("Есть \n", organization.GitlabUrl+repositoryRealisation)

					err = usecase.CloningRepository(organization.GitlabUrl+repositoryRealisation,
						organization.LocalPath+"/implementation/"+repositoryRealisation,
						&http.BasicAuth{
							Username: envopt.GetEnv("GITHUB_USER"),
							Password: envopt.GetEnv("GITLAB_TOKEN"),
						})

					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					//
					err := helpers.GitCheckoutDev(organization.LocalPath + "/implementation/" + repositoryRealisation)
					if err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					project.GitlabUrl = organization.GitlabUrl + repositoryRealisation

					s.DB.Save(&project)

				} else {
					fmt.Printf("Репозиторий %v не найден\n", organization.GitlabUrl+repositoryRealisation)

				}
			}

		}(project, organization)

	}
	wg.Wait()

	// клонирование репозиториев c реализацией

	ctx.JSON(200, organization)

}
