package service

import (
	"context"
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
	"os"
	"path/filepath"
)

func (s *Service) Organization(ctx *gin.Context) {

	organization := entity.Organization{}
	err := s.DB.Model(&organization).Where("name = ?", ctx.Param("name")).Take(&organization).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		ctx.HTML(200, "new_organization", gin.H{})
		ctx.Next()
	}

	projects := []entity.Project{}

	query := s.DB.Model(&entity.Project{})

	err = query.Where("organization_id = ?", organization.Id).Order("github_last_commit_time DESC").Limit(100).Find(&projects).Error

	if err != nil {

		fmt.Printf("Error: %v\n", err)
		return
	}

	repos, err := s.getOrganizationRepositories(organization.Name)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	for _, repo := range repos {

		hasDb := false
		for i, project := range projects {
			if project.Name == *repo.Name {
				hasDb = true
				if project.Type == entity.PROJECT_TYPE_REPOSITORY && project.PushedAt != repo.GetPushedAt().UTC() {

					release, _ := s.getLastRelease(organization.Name, *repo.Name)
					commit, _ := s.getLastCommit(organization.Name, *repo.Name)

					if release.GetTagName() != project.GithubReleaseTag {
						projects[i].NewTag = release.GetTagName()
					}

					projects[i].NewCommitName = commit.GetCommit().GetMessage()
					projects[i].NewCommitDate = commit.Commit.GetAuthor().GetDate()
				}
			}

			path := filepath.FromSlash("./tmp/" + project.Name)
			if _, err := os.Stat(path); err == nil {
				projects[i].HasClone = true
			}
		}

		if !hasDb {
			project := entity.Project{
				OrganizationId: organization.Id,
				Name:           *repo.Name,
				LocalPath:      *repo.FullName,
				Type:           1,
			}
			projects = append([]entity.Project{project}, projects...)
		}

	}

	//projects

	//commit, _ := s.getLastCommit(organization, *repos[0].Name, token)

	//byte, err := json.Marshal(commit)
	//log.Println(string(byte))
	ctx.HTML(200, "organization_list", gin.H{
		"Projects":     projects,
		"Organization": organization,
	})
	//SetPayload(ctx, viewData)
	ctx.Next()
}

func (s *Service) CreateOrganization(ctx *gin.Context) {

	ctx.HTML(200, "organization_create", gin.H{
		"Path": os.Getenv("GOPATH") + "/src/",
	})
}

func (s *Service) getOrganizationRepositories(organization string) ([]*github.Repository, error) {
	ctx := context.Background()

	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 10},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := s.GitClient.Repositories.ListByOrg(ctx, organization, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (s *Service) getRepository(owner, repo string) (*github.Repository, error) {
	ctx := context.Background()
	repository, _, err := s.GitClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func (s *Service) getLastCommit(owner, repoName string) (*github.RepositoryCommit, error) {
	ctx := context.Background()

	// Получение списка коммитов с лимитом 1.
	commits, _, err := s.GitClient.Repositories.ListCommits(ctx, owner, repoName, &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil {
		return nil, err
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found in the repository")
	}

	return commits[0], nil
}

func (s *Service) getLastRelease(owner, repoName string) (*github.RepositoryRelease, error) {
	ctx := context.Background()

	// Получение списка релизов с лимитом 1.
	releases, _, err := s.GitClient.Repositories.ListReleases(ctx, owner, repoName, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found in the repository")
	}

	return releases[0], nil
}
