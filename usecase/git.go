package usecase

import (
	"generator/entity"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"

	"os"
	"path/filepath"
)

func CloningRepository(gitRepository string, basicAuth *http.BasicAuth) (clonePath string) {

	clonePath = filepath.FromSlash("./tmp")
	RemoveContents(clonePath)

	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:      gitRepository,
		Progress: os.Stdout,
		Auth:     basicAuth,
	})

	if err != nil {

		log.Fatal(err.Error())
	}

	log.WithFields(log.Fields{
		"Status": "Complete",
	}).Info("Cloning  repository")

	return
}

func GetRepositoryInfo(funcFile string) (pack entity.PackageStruct)  {
	dat, err := ioutil.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)
	pack = ParseProtobufSourceAddress(source)
	return
}