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
	"unicode"
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

			byteSource, err := format.Source([]byte(code))
			if err != nil {
				fmt.Println("Error formatting code:", err)
				//return
			}
			formattedCodeNewCode := string(byteSource)

			saveFilePath := servicePath + "/entity/" + strcase.ToSnake(l.Name) + ".go"
			//if replaceFile {
			//	err = FileSave(saveFilePath, code)
			//	if err == nil {
			//		log.WithField("File", saveFilePath).Println("Entity created")
			//	}
			//}
			formattedCodeOldCode := ""
			hasFile := false
			hasDiff := true
			if _, err := os.Stat(saveFilePath); err == nil {
				hasFile = true

				file, err := os.ReadFile(saveFilePath)
				if err != nil {
					log.Fatalf("Ошибка при чтении файла: %v", err)
				}

				byteSource, err := format.Source(file)
				if err != nil {
					fmt.Println("Error formatting code:", err)
				}

				formattedCodeOldCode = string(byteSource)

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

func removeWhitespace(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, s)
}

func CompareStrings(s1, s2 string) bool {
	return removeWhitespace(s1) == removeWhitespace(s2)
}
