package service

import (
	"context"
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v39/github"
	"os"
	"path/filepath"
	"sync"
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

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, repo := range repos {
		hasDb := false
		for i, _ := range projects {
			wg.Add(1)
			go func(i int, repo *github.Repository, hasDb *bool) {
				defer wg.Done()

				project := projects[i]
				if project.Name == *repo.Name {

					mutex.Lock()
					*hasDb = true
					mutex.Unlock()
					if project.PushedAt != repo.GetPushedAt().UTC() {

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
			}(i, repo, &hasDb)
		}

		wg.Wait()

		if !hasDb {
			project := entity.Project{
				OrganizationId: organization.Id,
				Name:           *repo.Name,
				LocalPath:      *repo.FullName,
				Type:           entity.GetTypeProjectByName(*repo.Name),
				IsNewProject:   true,
			}
			projects = append([]entity.Project{project}, projects...)
		}
	}

	ctx.HTML(200, "organization_list", gin.H{
		"Projects":     projects,
		"Organization": organization,
	})
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
		repos, resp, err := s.GitHubClient.Repositories.ListByOrg(ctx, organization, opt)
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
	repository, _, err := s.GitHubClient.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	return repository, nil
}

func (s *Service) getLastCommit(owner, repoName string) (*github.RepositoryCommit, error) {
	ctx := context.Background()

	// Получение списка коммитов с лимитом 1.
	commits, _, err := s.GitHubClient.Repositories.ListCommits(ctx, owner, repoName, &github.CommitsListOptions{
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
	releases, _, err := s.GitHubClient.Repositories.ListReleases(ctx, owner, repoName, &github.ListOptions{PerPage: 1})
	if err != nil {
		return nil, err
	}

	if len(releases) == 0 {
		return nil, fmt.Errorf("no releases found in the repository")
	}

	return releases[0], nil
}
