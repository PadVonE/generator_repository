package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os/exec"
	"strings"
)

func (s *Service) GenerateProtobuf(ctx *gin.Context) {

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if project.Type != entity.PROJECT_TYPE_REPOSITORY && project.Type != entity.PROJECT_TYPE_USECASE {
		log.Warn("Not Repository or Usecase project.Id: %v\n", project.Id)
		return
	}

	organizationId := project.OrganizationId

	//	Поиск органицации в базе
	organization := entity.Organization{}
	err = s.DB.Model(&organization).Where("id = ?", organizationId).Take(&organization).Error
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	s.LogGlobal(" --- Go protobuf generator --- " + project.Name)

	protoPath := organization.LocalPath + "/proto"

	url := project.RealisationUrl
	url = strings.Replace(url, "https://", "", -1)
	repoPath := strings.Replace(url, "http://", "", -1)

	cmd := exec.Command(
		"docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/var/proto", protoPath),
		"--name", "proto", "proto-image",
		"/bin/bash", "-c",
		fmt.Sprintf("cd /var/proto && protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative %s/*.proto", repoPath),
	)

	cmd.Dir = protoPath

	realTimeOutput(cmd)

	s.LogComplete(" --- Complete Go protobuf generator --- " + project.Name)

}
