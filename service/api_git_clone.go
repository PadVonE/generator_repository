package service

import (
	"encoding/json"
	"generator/entity"
	"generator/usecase"
	"github.com/gin-gonic/gin"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (s *Service) GitClone(ctx *gin.Context) {

	repository:= "https://github.com/fontionis/proto-advertiser-repository"

	clonePath := s.cloneRepository(repository)


	// packageInfo - Хранит информацию о названии пакета, полного пути на гите, названия для импорти,
	// listOfStruct - список всех структур описанных в протофайлах (Название структуры,Тип структуры, Поля)
	// funcList - Структура, (в которой распарсеный файл *_repository.proto), Хранит список методов для реализации а также request и response структуры
	packageInfo, listOfStruct, funcList := s.parseInfoFromProto(clonePath)

	groupStructWithMethod(&funcList, listOfStruct)

	packageInfoJson,err :=  json.Marshal(packageInfo)
	if err != nil {
		log.Println(err)
	}

	listOfStructJson,err :=  json.Marshal(listOfStruct)
	if err != nil {
		log.Println(err)
	}

	funcListJson,err :=  json.Marshal(funcList)
	if err != nil {
		log.Println(err)
	}

	item := entity.Project{
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		GithubProto:  repository,
		FolderName:   "",
		FolderRepository:  clonePath,
		Type:         0,
		Version:      "",
		PackageInfo:  string(packageInfoJson),
		ListOfStruct: string(listOfStructJson),
		FuncList:     string(funcListJson),
	}

	_, err = s.DB.Insert(&item)

	if err != nil {
		log.Println(err)
	}

	ctx.JSON(200, packageInfo)
}


func  (s *Service) cloneRepository(repositoryLink string) (clonePath string,err error) {

	item := entity.AccessData{}

	query := s.DB.NewSession()

	has, err := query.Get(&item)
	if err != nil || !has {
		return clonePath,err
	}

	repositoryNameSpliced := strings.Split(repositoryLink, "/")

	clonePath = filepath.FromSlash("./tmp/"+repositoryNameSpliced[len(repositoryNameSpliced)-1])

	err = usecase.RemoveContents(clonePath)
	if err != nil  {
		return clonePath,err
	}


	_, err = git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:      repositoryLink,
		Progress: os.Stdout,
		Auth:    &http.BasicAuth{
			Username: item.GithubUsername,
			Password: item.GithubToken,
		},
	})

	return clonePath,err
}


func (s *Service) parseInfoFromProto(clonePath string) (packageInfo entity.PackageStruct, listOfStruct []entity.Struct, funcList entity.ProtoInterface) {

	files, err := ioutil.ReadDir(clonePath)
	if err != nil {
		log.Fatal(err)
	}

	funcFile := ""
	protoFiles := []string{}

	for _, file := range files {

		fileAddress := clonePath + "/" + file.Name()
		if strings.HasSuffix(file.Name(), "repository.pb.go") {
			funcFile = fileAddress
			continue
		}

		if strings.HasSuffix(file.Name(), ".pb.go") {
			protoFiles = append(protoFiles, fileAddress)
		}
	}

	packageInfo = usecase.GetRepositoryInfo(funcFile)

	listOfStruct = []entity.Struct{}

	for _, file := range protoFiles {

		dat, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		sourceFile := string(dat)
		listOfStruct = append(listOfStruct, usecase.ParseProtobufStruct(sourceFile)...)
	}

	dat, err := ioutil.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)

	funcList = usecase.ParseProtobufFunc(source)

	return
}

func groupStructWithMethod(funcList *entity.ProtoInterface, listOfStruct []entity.Struct) {

	for i, f := range funcList.Methods {

		for _, los := range listOfStruct {
			if f.Request == los.Name {
				funcList.Methods[i].RequestStruct = los
			}
			if f.Response == los.Name {
				funcList.Methods[i].ResponseStruct = los
			}

			if f.Basic == los.Name {
				funcList.Methods[i].BasicStruct = los
			}
		}
	}
}
