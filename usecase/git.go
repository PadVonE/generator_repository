package usecase

import (
	"generator/entity"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
)

func CloningRepository(gitRepository string, basicAuth *http.BasicAuth) (clonePath string) {



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