package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
)

func (s *Service) ProjectListApi(ctx *gin.Context) {
	type FilterResponse struct {
		Id   int32  `json:"id"`
		Name string `json:"name"`
	}
	organizationId := ctx.Query("organization_id")

	// Получаем список проектов
	projects := []entity.Project{}

	query := s.DB.Model(&entity.Project{})

	err := query.Where("organization_id = ?", organizationId).Order("name ASC").Find(&projects).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	mapProject := map[string][]FilterResponse{}

	mapProject["usecase"] = []FilterResponse{}
	mapProject["repository"] = []FilterResponse{}
	mapProject["gateway"] = []FilterResponse{}
	mapProject["other"] = []FilterResponse{}

	for _, project := range projects {

		switch project.Type {
		case entity.PROJECT_TYPE_REPOSITORY:
			mapProject["repository"] = append(mapProject["repository"], FilterResponse{Id: project.Id, Name: project.GetRealisationName()})
		case entity.PROJECT_TYPE_USECASE:
			mapProject["usecase"] = append(mapProject["usecase"], FilterResponse{Id: project.Id, Name: project.GetRealisationName()})
		case entity.PROJECT_TYPE_SPECIFICATION:
			mapProject["gateway"] = append(mapProject["gateway"], FilterResponse{Id: project.Id, Name: project.GetRealisationName()})
		default:
			mapProject["other"] = append(mapProject["other"], FilterResponse{Id: project.Id, Name: project.GetRealisationName()})
		}

	}

	ctx.JSON(200, gin.H{
		"projects": mapProject,
		"err":      err,
	})

}
