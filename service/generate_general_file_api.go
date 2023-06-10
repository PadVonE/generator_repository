package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators/repository"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) GenerateGeneralFileApi(ctx *gin.Context) {
	log.Println("\033[35m", "\n\nGeneral Files file", "\033[0m")

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	servicePath := filepath.FromSlash(project.LocalPath)

	projectComponents := entity.ProjectComponents{}
	err = json.Unmarshal([]byte(project.LastStructure), &projectComponents)

	type GeneralFile struct {
		FileName string
		Replace  bool
	}

	listFiles := []GeneralFile{
		{".gitignore", false},
		{"db.go", false},
		{"envopt.json", false},
		{"go.mod", false},
		{"go.sum", false},
		{"main.go", false},
		{"server.go", false},
		{"service/service.go", false},
		{"healthcheck/healthcheck.go", false},
	}

	//if isGenerateTestFile {
	listFiles = append(listFiles, GeneralFile{"service/service_test.go", true})
	listFiles = append(listFiles, GeneralFile{"envopt_test.json", false})
	//}

	dbList := []string{}
	for _, l := range projectComponents.ListOfStruct {
		if l.Type == entity.TypeMain {
			dbList = append(dbList, strcase.ToSnake(l.Name))
		}
	}

	response := []entity.FilesPreview{}

	for _, l := range listFiles {
		saveFilePath := servicePath + "/" + l.FileName
		// Проверка на то что файл не существует
		//if !l.Replace {
		//	if _, err := os.Stat(saveFilePath); err == nil {
		//		continue
		//	}
		//}

		code, err := generators.GenerateGeneral(l.FileName, projectComponents.PackageStruct, dbList)
		if err != nil {
			log.Error(err)
			continue
		}

		formattedCodeNewCode := code

		if strings.HasSuffix(saveFilePath, ".go") {
			byteSource, err := format.Source([]byte(code))
			if err != nil {
				fmt.Println("Error formatting code:", err)
				return
			}
			formattedCodeNewCode = string(byteSource)
		}

		formattedCodeOldCode := ""
		hasFile := false
		hasDiff := true
		if _, err := os.Stat(saveFilePath); err == nil {
			hasFile = true

			file, err := os.ReadFile(saveFilePath)
			if err != nil {
				log.Fatalf("Ошибка при чтении файла: %v", err)
			}

			formattedCodeOldCode = string(file)
			if strings.HasSuffix(saveFilePath, ".go") {
				byteSource, err := format.Source(file)
				if err != nil {
					fmt.Println("Error formatting code:", err)
					return
				}
				formattedCodeOldCode = string(byteSource)
			}

			if CompareStrings(formattedCodeOldCode, formattedCodeNewCode) {
				hasDiff = false
			}
		}

		response = append(response, entity.FilesPreview{
			FilePath: saveFilePath,
			NewCode:  formattedCodeNewCode,
			OldCode:  formattedCodeOldCode,
			HasFile:  hasFile,
			HasDiff:  hasDiff,
		})

	}

	ctx.JSON(200, response)

}
