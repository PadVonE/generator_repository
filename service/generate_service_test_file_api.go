package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"go/format"
	"os"
	"path/filepath"
)

func (s *Service) GenerateServiceTestFileApi(ctx *gin.Context) {
	log.Println("\033[35m", "\n\nTEST files", "\033[0m")

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
	response := []entity.FilesPreview{}

	for _, pi := range projectComponents.ListOfFunction.Methods {
		// Generate file

		nameInterface := pi.NameInterface(&projectComponents.ListOfFunction)
		saveFilePath := servicePath + "/service/" + nameInterface.FileName() + "_test.go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(nameInterface.Action) == 0 {
			continue
		}

		// Generate tests

		codeTest := ""
		switch nameInterface.Action {
		case "Create":
			codeTest, err = generators.GenerateTestCreateCode(pi, projectComponents.PackageStruct, nameInterface)
		case "Update":
			codeTest, err = generators.GenerateTestUpdateCode(pi, projectComponents.PackageStruct, nameInterface)
		case "Delete":
			codeTest, err = generators.GenerateTestDeleteCode(pi, projectComponents.PackageStruct, nameInterface)
		case "Get":
			codeTest, err = generators.GenerateTestGetCode(pi, projectComponents.PackageStruct, nameInterface)
		case "List":
			codeTest, err = generators.GenerateTestListCode(pi, projectComponents.PackageStruct, nameInterface)

		}

		if len(codeTest) != 0 {

			if err != nil {
				log.Error(err)
				continue
			}

			byte, err := format.Source([]byte(codeTest))
			if err != nil {
				fmt.Println("Error formatting code:", err)
				return
			}
			formattedCodeNewCode := string(byte)

			formattedCodeOldCode := ""
			hasFile := false
			hasDiff := true
			if _, err := os.Stat(saveFilePath); err == nil {
				hasFile = true

				file, err := os.ReadFile(saveFilePath)
				if err != nil {
					log.Fatalf("Ошибка при чтении файла: %v", err)
				}

				byte, err := format.Source(file)
				if err != nil {
					fmt.Println("Error formatting code:", err)
					return
				}

				formattedCodeOldCode = string(byte)

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

	}

	ctx.JSON(200, response)

}
