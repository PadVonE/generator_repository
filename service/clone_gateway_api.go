package service

import (
	"encoding/json"
	"fmt"
	"generator/entity"
	"generator/usecase"
	"github.com/2q4t-plutus/envopt"
	"github.com/gin-gonic/gin"
	"github.com/go-openapi/loads"
	"github.com/go-openapi/spec"
	log "github.com/sirupsen/logrus"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"path/filepath"
)

func (s *Service) CloneGatewayApi(ctx *gin.Context) {
	projectID := ctx.Query("project_id")

	specificationProjectComponents, err := s.CloneGateway(projectID)

	ctx.JSON(200, gin.H{
		"struct": specificationProjectComponents,
		"err":    err,
	})

}

func (s *Service) CloneGateway(projectID string) (specificationProjectComponents entity.SpecificationProjectComponents, err error) {
	project := entity.Project{}

	query := s.DB.Model(&project)

	err = query.Where("id = ?", projectID).Take(&project).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if project.Id == 0 {
		fmt.Printf("project.Id: %v\n", project.Id)
		return
	}

	if project.Type != entity.PROJECT_TYPE_SPECIFICATION {
		fmt.Printf("Not Specefication project.Id: %v\n", project.Id)
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

	swaggerSpec, err := loads.Spec(clonePath + "/main.yaml")
	if err != nil {
		log.Error("Failed to load swagger spec: %v", err)
	}

	// Expand spec to resolve all $ref
	swaggerSpec, err = swaggerSpec.Expanded()
	if err != nil {
		log.Error("Failed to expand swagger spec: %v", err)
	}

	info := swaggerSpec.Spec().Info
	specificationProjectComponents = entity.SpecificationProjectComponents{
		Name:    info.Title,
		Version: info.Version,
	}

	specificationProjectComponents.Path = AnalyzeSpec(swaggerSpec.Spec())

	jsonStruct, err := json.Marshal(specificationProjectComponents)

	cloneHistory := entity.CloneHistory{
		ProjectId:   project.Id,
		Name:        "",
		CloningPath: clonePath,
		ReleaseTag:  "",
		Structure:   string(jsonStruct),
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
		project.NewTag = release.GetTagName()
	}

	project.NewCommitName = commit.GetCommit().GetMessage()
	project.NewCommitDate = commit.Commit.GetAuthor().GetDate()

	project.LastStructure = string(jsonStruct)
	err = s.DB.Save(&project).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	return
}

func AnalyzeSpec(swaggerSpec *spec.Swagger) []entity.PathInfo {
	var pathsInfo []entity.PathInfo

	for path, pathItem := range swaggerSpec.Paths.Paths {
		pathInfo := entity.PathInfo{
			Path:       path,
			Operations: analyzeOperations(&pathItem),
		}
		pathsInfo = append(pathsInfo, pathInfo)
	}

	return pathsInfo
}

func analyzeOperations(pathItem *spec.PathItem) []entity.OperationInfo {
	var operationsInfo []entity.OperationInfo

	operations := []*spec.Operation{
		pathItem.Get, pathItem.Put, pathItem.Post, pathItem.Delete, pathItem.Options, pathItem.Head, pathItem.Patch,
	}

	for _, operation := range operations {
		if operation != nil {
			opInfo := entity.OperationInfo{
				NameMethod: operation.ID,
				Request:    getRequestProperties(operation.Parameters),
				Responses:  getResponseProperties(operation.Responses.StatusCodeResponses),
				Tag:        operation.Tags[0],
			}
			operationsInfo = append(operationsInfo, opInfo)
		}
	}

	return operationsInfo
}

func getRequestProperties(parameters []spec.Parameter) []entity.Property {
	var requestProps []entity.Property

	for _, param := range parameters {
		if param.Schema != nil {
			props := getSchemaProperties(param.Schema, "")
			requestProps = append(requestProps, props...)
		} else {
			prop := entity.Property{
				Name: param.Name,
				Type: param.Type,
			}
			requestProps = append(requestProps, prop)
		}
	}

	return requestProps
}

func getResponseProperties(statusCodeResponses map[int]spec.Response) map[int][]entity.Property {
	responseProps := make(map[int][]entity.Property)

	for statusCode, response := range statusCodeResponses {
		props := getSchemaProperties(response.Schema, "")
		responseProps[statusCode] = props
	}

	return responseProps
}

func getSchemaProperties(schema *spec.Schema, indent string) []entity.Property {
	var properties []entity.Property

	if schema == nil {
		return properties
	}

	for propName, propSchema := range schema.Properties {
		nameType := ""
		if len(propSchema.Type) > 0 {
			nameType = propSchema.Type[0]
		}
		child := getSchemaProperties(&propSchema, indent+"  ")
		if nameType == "array" {
			child = getSchemaProperties(propSchema.Items.Schema, indent+"  ")
		}

		prop := entity.Property{
			Name:     propName,
			Type:     nameType,
			Children: child,
		}

		properties = append(properties, prop)
	}

	return properties
}

func printProperties(properties []entity.Property, indent string) {
	for _, prop := range properties {
		fmt.Printf("%s%s: %s\n", indent, prop.Name, prop.Type)
		if len(prop.Children) > 0 {
			printProperties(prop.Children, indent+"  ")
		}
	}
}
