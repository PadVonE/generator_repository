package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	generators "generator/generators/gateway"
	"github.com/gin-gonic/gin"
	"github.com/iancoleman/strcase"
	log "github.com/sirupsen/logrus"
	"go/format"
	"os"
	"path/filepath"
	"strings"
)

func (s *Service) GenerateGatewayFileApi(ctx *gin.Context) {
	log.Info("Gateway files")

	//prefixList := []string{"List"}
	prefixList := []string{"List", "Get", "Create", "Update", "Delete", "Ping"}

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	s.LogGlobal(" --- Gateway files --- " + project.Name)
	org := entity.Organization{}

	query := s.DB.Model(&org)

	err = query.Where("id = ?", project.OrganizationId).Take(&org).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	servicePath := filepath.FromSlash(project.LocalPath)

	// Парсим данные структуры которые нужны для генирации кода
	projectComponents := entity.SpecificationProjectComponents{}
	err = json.Unmarshal([]byte(project.LastStructure), &projectComponents)

	// Получаем все проекты
	allRepository := []entity.Project{}
	err = s.DB.Find(&allRepository, "type IN ? AND last_structure != ?", []int{
		entity.PROJECT_TYPE_REPOSITORY,
		entity.PROJECT_TYPE_USECASE,
	}, "{}").Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	// Парсим все проекты в поисках нужных нам методов
	err = ParseProjects(allRepository)
	if err != nil {
		log.Error("Failed to parse projects: %v", err)
	}

	gatewayAction := ""
	gatewayName := ""
	response := []entity.FilesPreview{}

	for _, pc := range projectComponents.Path {

		for _, operation := range pc.Operations {

			relatedProject, err := FindProjectsByOperations(allRepository, operation.NameMethod)
			if relatedProject.Id == 0 {
				log.Errorf("Failed to find project for operation %s: %v", operation.NameMethod, err)
				continue
			}

			log.Infof("Found project %s for operation %s\n", relatedProject.Name, operation.NameMethod)

			for _, prefix := range prefixList {
				if strings.HasPrefix(operation.NameMethod, prefix) {
					gatewayAction = prefix
					gatewayName = strings.Replace(operation.NameMethod, prefix, "", 1)
				}
			}
			if len(gatewayAction) == 0 {
				log.Warn("Not found Action in name method " + operation.NameMethod)
				//continue
			}

			nameServiceGateway := "gateway-" + strings.TrimPrefix(project.Name, "specification-")

			code, err := generators.GenerateGatewayCode(&operation, gatewayName, gatewayAction, nameServiceGateway, org, relatedProject)

			if err != nil {
				log.Error(err)
				continue
			}

			byteSource, err := format.Source([]byte(code))
			if err != nil {
				log.Error("Error formatting code:", err)
				return
			}
			formattedCodeNewCode := string(byteSource)

			formattedCodeOldCode := ""
			hasFile := false
			hasDiff := true

			saveFilePath := servicePath + "/service/" + strcase.ToSnake(gatewayName) + "_" + strcase.ToSnake(gatewayAction) + ".go"

			if _, err := os.Stat(saveFilePath); err == nil {
				hasFile = true

				file, err := os.ReadFile(saveFilePath)
				if err != nil {
					log.Errorf("Ошибка при чтении файла: %v", err)
				}

				byteSource, err := format.Source(file)
				if err != nil {
					log.Error("Error formatting code:", err)
					return
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

func ParseProjects(allRepository []entity.Project) error {

	for i, project := range allRepository {
		err := json.Unmarshal([]byte(project.LastStructure), &allRepository[i].LastStructureUnmarshal)
		if err != nil {
			return fmt.Errorf("failed to unmarshal LastStructure for project ID %d: %v", project.Id, err)
		}
	}

	return nil
}

func FindProjectsByOperations(allProject []entity.Project, operation string) (projectOperation entity.Project, err error) {

	for _, project := range allProject {
		for _, structObj := range project.LastStructureUnmarshal.ListOfFunction.Methods {
			if structObj.NameMethod == operation {
				projectOperation = project
				break
			}
		}
	}

	return
}
