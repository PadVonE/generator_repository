package service

import (
	"fmt"
	"generator/entity"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
)

func (s *Service) GenerateSwaggerServer(ctx *gin.Context) {

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	s.LogGlobal(" --- Go swagger generator --- " + project.Name)

	specificationPath := filepath.FromSlash("./tmp/" + project.Name + "/main.yaml")
	outputPath := project.LocalPath

	// Получение абсолютного пути к файлам спецификации
	specificationPathAbs, err := filepath.Abs(specificationPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for specification files: %s", err)
	}

	// Получение абсолютного пути к папке для вывода
	outputPathAbs, err := filepath.Abs(outputPath)
	if err != nil {
		log.Fatalf("Failed to get absolute path for output folder: %s", err)
	}

	// Проверка, существует ли путь к спецификациям
	if _, err := os.Stat(specificationPathAbs); os.IsNotExist(err) {
		log.Fatalf("Specification path does not exist: %s", specificationPathAbs)
	}

	// Проверка, существует ли путь для вывода
	if _, err := os.Stat(outputPathAbs); os.IsNotExist(err) {
		log.Fatalf("Output path does not exist: %s", outputPathAbs)
	}

	// Выполнение команды swagger через docker
	cmd := exec.Command("docker", "run",
		"--rm",
		"--user", fmt.Sprintf("%d:%d", os.Getuid(), os.Getgid()),
		"-e", fmt.Sprintf("GOPATH=%s:/go", os.Getenv("GOPATH")),
		"-v", fmt.Sprintf("%s:%s", os.Getenv("HOME"), os.Getenv("HOME")),
		"-w", outputPathAbs,
		"quay.io/goswagger/swagger",
		"generate", "server", "--with-flatten=full", "-f", specificationPathAbs)

	realTimeOutput(cmd)

	log.Infof("Successfully generated swagger server in %s", outputPath)

	// Запуск go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = outputPathAbs // Указание рабочего каталога для выполнения команды
	realTimeOutput(cmd)

	log.Info("Successfully tidied go modules")

}
