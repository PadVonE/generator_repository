package helpers

import (
	"github.com/xanzy/go-gitlab"
)

func RepoExists(token string, repoName string) (bool, error) {
	gitlabClient := gitlab.NewClient(nil, token)

	_, _, err := gitlabClient.Projects.GetProject(repoName, nil)
	if err != nil {
		if errResponse, ok := err.(*gitlab.ErrorResponse); ok && errResponse.Response.StatusCode == 404 {
			return false, nil
		}

		return false, err
	}

	return true, nil
}
