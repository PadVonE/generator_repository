package service

import (
	"bufio"
	"generator/entity"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
)

func (s *Service) SshUpdateAndTidyModules(ctx *gin.Context) {

	projectID := ctx.Query("project_id")

	project := entity.Project{}

	err := s.DB.First(&project, "id = ?", projectID).Error

	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	s.LogGlobal(" --- Go Get -U AND GO mod tidy --- " + project.Name)

	servicePath := project.LocalPath

	if project.Type == entity.PROJECT_TYPE_SPECIFICATION {
		newPath := filepath.Join(servicePath, "cmd")

		files, err := os.ReadDir(newPath)
		if err != nil {
			log.Error(err)
			return
		}

		for _, file := range files {
			if file.IsDir() {
				servicePath = filepath.Join(newPath, file.Name())
				break
			}
		}
	}

	// Проверка, существует ли путь
	if _, err := os.Stat(servicePath); os.IsNotExist(err) {
		log.Errorf("Service path does not exist: %s", servicePath)
		return
	}

	// Выполнение команды go get -u
	cmd := exec.Command("go", "get", "-u")
	cmd.Dir = servicePath
	realTimeOutput(cmd)

	// Выполнение команды go mod tidy
	cmd = exec.Command("go", "mod", "tidy")
	cmd.Dir = servicePath
	realTimeOutput(cmd)

	log.Infof("Successfully updated and tidied Go modules in %s", servicePath)
}

func realTimeOutput(cmd *exec.Cmd) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
		wg.Done()
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Info(scanner.Text())
		}
		wg.Done()
	}()

	err := cmd.Start()
	if err != nil {
		log.Error(err)
	}

	wg.Wait()
	err = cmd.Wait()

	if err != nil {
		log.Error(err)
	}
}
