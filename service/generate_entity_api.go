package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/generators"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"go/format"
	"os"
	"path/filepath"
	"strings"
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

			byte, err := format.Source([]byte(code))
			if err != nil {
				fmt.Println("Error formatting code:", err)
				return
			}
			formattedCodeNewCode := string(byte)

			saveFilePath := servicePath + "/entity/" + strcase.ToSnake(l.Name) + ".go"
			//if replaceFile {
			//	err = FileSave(saveFilePath, code)
			//	if err == nil {
			//		log.WithField("File", saveFilePath).Println("Entity created")
			//	}
			//}
			formattedCodeOldCode := ""
			hasFile := false
			hasDiff := false
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

				if removeSpacesAndNewlines(formattedCodeOldCode) == removeSpacesAndNewlines(formattedCodeNewCode) {
					hasDiff = true
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

func removeSpacesAndNewlines(s string) string {
	withoutSpaces := strings.ReplaceAll(s, " ", "")
	withoutEnter := strings.ReplaceAll(withoutSpaces, "\n", "")
	withoutNewlines := strings.ReplaceAll(withoutEnter, "\t", "")
	return withoutNewlines
}
