package usecase

import (
	"generator/entity"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"os"
)

func CloningRepository(gitRepository string, clonePath string, basicAuth *http.BasicAuth) error {

	_ = RemoveContents(clonePath)

	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:      gitRepository,
		Progress: os.Stdout,
		Auth:     basicAuth,
	})

	if err != nil {
		return err
	}

	//log.WithFields(log.Fields{
	//	"Status": "Complete",
	//}).Info("Cloning  repository")

	return nil
}

func GetRepositoryInfo(funcFile string) (pack entity.PackageStruct) {
	dat, err := os.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)
	pack = ParseProtobufSourceAddress(source)
	return
}
