package service

import (
	"fmt"
	"generator/entity"
	"generator/helpers"
	"generator/usecase"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
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
		go func(wg *sync.WaitGroup, repo *github.Repository, organization *entity.Organization, projects []entity.Project) {
			defer wg.Done()
			hasDb := false
			for _, project := range projects {
				if project.Name == *repo.Name {
					hasDb = true
				}
			}

			if !hasDb {
				s.createProject(repo, organization)
			}

		}(&wg, repo, &organization, projects)
	}
	wg.Wait()

	// После добавления нужно перезапросить все репозитории
	query = s.DB.Model(&entity.Project{})
	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error

	// Добавляем реализации если нет спек
	projectURL := strings.ReplaceAll(organization.GitlabUrl, "https://gitlab.com/", "")
	projectURL = strings.TrimSuffix(projectURL, "/")

	gitlabRepos, _, err := s.GitLabClient.Groups.ListGroupProjects(projectURL, nil)
	if err != nil {
		panic(err)
	}

	for _, repo := range gitlabRepos {
		wg.Add(1)
		go func(wg *sync.WaitGroup, repo *gitlab.Project, organization *entity.Organization, projects []entity.Project) {
			defer wg.Done()
			hasDb := false
			for _, project := range projects {
				projectName := strings.ReplaceAll(project.Name, "proto-", "")
				projectName = strings.ReplaceAll(projectName, "specification-", "gateway-")

				if projectName == repo.Name {
					hasDb = true
				}
			}

			if !hasDb {
				log.Error(" repo.Name ", repo.Name)
				s.createProjectByRealisation(repo, organization)
			}

		}(&wg, repo, &organization, projects)
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
			log.Error("Ошибка при создании папок: %v\n", err)
			return
		}
	}

	// клонирование репозиториев proto в папку proto

	projects = []entity.Project{}

	query = s.DB.Model(&entity.Project{})
	err = query.Where("organization_id = ?", organization.Id).Find(&projects).Error
	for _, project := range projects {
		wg.Add(1)
		go func(wg *sync.WaitGroup, project entity.Project, organization entity.Organization) {
			defer wg.Done()

			s.cloneProtoAndRealisation(project, organization)

		}(&wg, project, organization)

	}
	wg.Wait()

	ctx.JSON(200, organization)

}

func (s *Service) createProject(repo *github.Repository, organization *entity.Organization) (project entity.Project) {

	commit, _ := s.getLastCommit(organization.Name, *repo.Name)
	release, _ := s.getLastRelease(organization.Name, *repo.Name)

	repositoryRealisation := entity.GetRealisationName(entity.GetTypeProjectByName(*repo.Name), repo.GetName())

	project = entity.Project{
		Type:                          entity.GetTypeProjectByName(*repo.Name),
		OrganizationId:                organization.Id,
		Name:                          repo.GetName(),
		LocalPath:                     organization.LocalPath + "/implementation/" + repositoryRealisation,
		SpecificationUrl:              "https://github.com/" + organization.Name + "/" + repo.GetName(),
		PushedAt:                      repo.GetPushedAt().UTC(),
		SpecificationLastCommitName:   commit.Commit.GetMessage(),
		SpecificationLastCommitTime:   repo.GetPushedAt().UTC(),
		SpecificationLastCommitAuthor: commit.Commit.GetAuthor().GetName(),
		SpecificationReleaseTag:       release.GetTagName(),
		LastStructure:                 "{}",
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

	s.DB.Create(&project)
	return

}

func (s *Service) createProjectByRealisation(realisationRepo *gitlab.Project, organization *entity.Organization) (project entity.Project) {

	project = entity.Project{
		Type:           entity.GetTypeProjectByName(realisationRepo.Name),
		OrganizationId: organization.Id,
		Name:           realisationRepo.Name,
		LocalPath:      organization.LocalPath + "/implementation/" + realisationRepo.Name,
		LastStructure:  "{}",
	}

	release, _ := s.getLastReleaseGitlab(realisationRepo)

	if release != nil && release.Name !=
		project.RealisationReleaseTag {
		project.RealisationReleaseTag = release.Name
	}

	commit, _ := s.getLastCommitGitlab(realisationRepo, "dev")
	if commit != nil {
		project.RealisationLastCommitName = commit.Title
		project.RealisationLastCommitAuthor = commit.AuthorName
	}
	project.RealisationLastCommitTime = *realisationRepo.LastActivityAt
	project.RealisationUrl = organization.GitlabUrl + realisationRepo.Name

	s.DB.Create(&project)
	return

}

func (s *Service) cloneProtoAndRealisation(project entity.Project, organization entity.Organization) {

	clonePath := entity.GetPath(project.Type, project.Name, &organization)
	repositoryRealisation := entity.GetRealisationName(project.Type, project.Name)

	log.Println(project.SpecificationUrl, clonePath, repositoryRealisation)
	if len(clonePath) > 0 {
		err := usecase.CloningRepository(project.SpecificationUrl,
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
		exist := true

		_, _, err := s.GitLabClient.Projects.GetProject(organization.GitlabOrganizationName()+"/"+repositoryRealisation, nil)
		if err != nil {
			if errResponse, ok := err.(*gitlab.ErrorResponse); ok && errResponse.Response.StatusCode == 404 {
				exist = false
			}
			exist = false
		}

		if exist {
			fmt.Printf("Есть \n", organization.GitlabUrl+repositoryRealisation)

			err = usecase.CloningRepository(organization.GitlabUrl+repositoryRealisation,
				organization.LocalPath+"/implementation/"+repositoryRealisation,
				&http.BasicAuth{
					Username: envopt.GetEnv("GITHUB_USER"),
					Password: envopt.GetEnv("GITLAB_TOKEN"),
				})

			if err != nil {
				log.Errorf("Error: %v\n %s", err, organization.GitlabUrl+repositoryRealisation)
			}

			err := helpers.GitCheckoutBranch(organization.LocalPath+"/implementation/"+repositoryRealisation, "dev")
			if err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		} else {
			fmt.Printf("Репозиторий %v не найден\n", organization.GitlabUrl+repositoryRealisation)

		}
	}
	return

}
