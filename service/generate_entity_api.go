package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

func (s *Service) GenerateEntityApi(ctx *gin.Context) {

	log.Println("\033[35m", "\n\nEntity files", "\033[0m")

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

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	response := []entity.FilesPreview{}

	for _, l := range projectComponents.ListOfStruct {
		if l.Type == entity.TypeMain {

			createFunction := false
			updateFunction := false

			// Определяем нужны ли функиции для создания и обновления даннх
			for _, tempStruct := range projectComponents.ListOfStruct {
				if tempStruct.Name == "Create"+l.Name+"Request" {
					createFunction = true
				}
				if tempStruct.Name == "Update"+l.Name+"Request" {
					updateFunction = true
				}
			}

			code, err := generators.GenerateEntity(l, projectComponents.PackageStruct, createFunction, updateFunction)
			if err != nil {
				log.Error(err)
				continue
			}

			saveFilePath := servicePath + "/entity/" + strcase.ToSnake(l.Name) + ".go"
			//if replaceFile {
			//	err = FileSave(saveFilePath, code)
			//	if err == nil {
			//		log.WithField("File", saveFilePath).Println("Entity created")
			//	}
			//}

			response = append(response, entity.FilesPreview{
				FilePath: saveFilePath,
				NewCode:  code,
				OldCode:  "",
				HasFile:  false,
			})
		}
	}

	ctx.JSON(200, response)

}
