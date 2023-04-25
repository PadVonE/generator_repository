package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func (s *Service) GenerateServiceFileApi(ctx *gin.Context) {
	log.Println("\033[35m", "\n\nService files", "\033[0m")

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

		code := ""
		nameInterface := pi.NameInterface(&projectComponents.ListOfFunction)
		saveFilePath := servicePath + "/service/" + nameInterface.FileName() + ".go"

		// Если не удалось определить экшн то переходим к следующему методу
		if len(nameInterface.Action) == 0 {
			log.Warn("action not allowed:", pi.NameMethod)
			continue
		}

		code, err = generators.GenerateServiceCode(pi, projectComponents.PackageStruct, nameInterface)

		if err != nil {
			log.Error(err)
			continue
		}

		response = append(response, entity.FilesPreview{
			FilePath: saveFilePath,
			NewCode:  code,
			OldCode:  "",
			HasFile:  false,
		})

	}

	ctx.JSON(200, response)

}
