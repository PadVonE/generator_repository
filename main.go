package main

import (
	"generator/entity"
	"generator/usecase"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"io/ioutil"
	"os"
	"strings"
)

const GitHubUsername = ""
const GitHubToken = ""
const GitRepository = "https://github.com/******/****"

const ServiceName = "teaser-repository"

const IsGenerateEntity = true
const IsGenerateServiceFile = true
const IsGenerateTestFile = true
const IsGenerateMigrationFile = false

func init() {
	// Log as JSON instead of the default ASCII formatter.
	log.SetFormatter(&log.TextFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.InfoLevel)
}

func main() {

	clonePath := usecase.CloningRepository(GitRepository,
		&http.BasicAuth{
			Username: GitHubUsername, // yes, this can be anything except an empty string
			Password: GitHubToken,
		})

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

	packageInfo := usecase.GetRepositoryInfo(funcFile)

	listOfStruct := []entity.Struct{}

	for _, file := range protoFiles {

		dat, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		sourceFile := string(dat)
		listOfStruct = append(listOfStruct, usecase.ParseProtobufStruct(sourceFile)...)
	}

	if IsGenerateEntity {
		usecase.GenerateEntity(packageInfo, ServiceName, listOfStruct)
	}

	if IsGenerateMigrationFile {
		usecase.GenerateMigrationFile(packageInfo, ServiceName, listOfStruct)
	}

	dat, err := ioutil.ReadFile(funcFile)
	if err != nil {
		panic(err)
	}
	source := string(dat)

	funcList := usecase.ParseProtobufFunc(source)

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

	if IsGenerateServiceFile {
		usecase.GenerateServiceFiles(packageInfo, funcList, ServiceName)
	}

	if IsGenerateTestFile {
		usecase.GenerateTestFiles(packageInfo, funcList, ServiceName)
	}

}
