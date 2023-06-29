package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/usecase"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) CloneRepositoryApi(ctx *gin.Context) {
	projectID := ctx.Query("project_id")

	project := entity.Project{}

	query := s.DB.Model(&project)

	err := query.Where("id = ?", projectID).Take(&project).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if project.Id == 0 {
		fmt.Printf("project.Id: %v\n", project.Id)
		return
	}
	if project.Type != entity.PROJECT_TYPE_REPOSITORY && project.Type != entity.PROJECT_TYPE_USECASE {
		fmt.Printf("Not Repository or Usecase project.Id: %v\n", project.Id)
		return
	}

	clonePath := filepath.FromSlash("./tmp/" + project.Name)

	err = usecase.CloningRepository(project.GithubUrl,
		clonePath,
		&http.BasicAuth{
			Username: envopt.GetEnv("GITHUB_USER"),
			Password: envopt.GetEnv("GITHUB_TOKEN"),
		})

	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Парсим информацию из прото файлов в нужные нам структуры
	projectComponents := ParseInfoFromProto(clonePath)

	// Групируем Структуры с методами в которых они вызываются
	GroupStructWithMethod(&projectComponents.ListOfFunction, projectComponents.ListOfStruct)

	structure, err := json.Marshal(projectComponents)

	cloneHistory := entity.CloneHistory{
		ProjectId:   project.Id,
		Name:        "",
		CloningPath: clonePath,
		ReleaseTag:  "",
		Structure:   string(structure),
	}

	err = s.DB.Create(&cloneHistory).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	org := entity.Organization{}

	query = s.DB.Model(&org)

	err = query.Where("id = ?", project.OrganizationId).Take(&org).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	repo, err := s.getRepository(org.Name, project.Name)

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	project.PushedAt = repo.GetPushedAt().UTC()

	release, _ := s.getLastRelease(org.Name, *repo.Name)
	commit, _ := s.getLastCommit(org.Name, *repo.Name)

	if release.GetTagName() != project.GithubReleaseTag {
		project.GithubReleaseTag = release.GetTagName()
	}

	project.GithubLastCommitAuthor = commit.GetCommit().GetAuthor().GetName()
	project.GithubLastCommitName = commit.GetCommit().GetMessage()
	project.GithubLastCommitTime = commit.Commit.GetAuthor().GetDate()

	project.LastStructure = string(structure)

	err = s.DB.Save(&project).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
}

func ParseInfoFromProto(clonePath string) (projectComponents entity.ProjectComponents) {

	files, err := os.ReadDir(clonePath)
	if err != nil {
		log.Fatal(err)
	}

	funcFile := ""
	protoFiles := []string{}

	for _, file := range files {

		fileAddress := clonePath + "/" + file.Name()
		if strings.HasSuffix(file.Name(), "repository.pb.go") ||
			strings.HasSuffix(file.Name(), "repository_grpc.pb.go") ||
			strings.HasSuffix(file.Name(), "usecase.pb.go") ||
			strings.HasSuffix(file.Name(), "usecase_grpc.pb.go") {
			funcFile = fileAddress
			continue
		}

		if strings.HasSuffix(file.Name(), ".pb.go") {
			protoFiles = append(protoFiles, fileAddress)
		}
	}

	log.Println(funcFile)
	projectComponents.PackageStruct = usecase.GetRepositoryInfo(funcFile)

	projectComponents.ListOfStruct = []entity.Struct{}

	for _, file := range protoFiles {

		dat, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		sourceFile := string(dat)
		projectComponents.ListOfStruct = append(projectComponents.ListOfStruct, usecase.ParseProtobufStruct(sourceFile)...)
	}

	dat, err := os.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)

	projectComponents.ListOfFunction = usecase.ParseProtobufFunc(source)

	return
}

func GroupStructWithMethod(funcList *entity.ProtoInterface, listOfStruct []entity.Struct) {

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
