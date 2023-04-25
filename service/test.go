package service

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"os"
	"os/exec"
)

func (s *Service) UpdateGoPackagesInDir(ctx *gin.Context) {
	//projectID := ctx.Query("project_id")
	//
	//project := entity.Project{}
	//
	//err := s.DB.First(&project, "id = ?", projectID).Error
	//
	//if err != nil {
	//	fmt.Printf("Error: %v\n", err)
	//	return
	//}
	//
	//servicePath := filepath.FromSlash(project.LocalPath)

	servicePath := "/Users/padvone/go/src/plutus/implementation/account-repository"
	err := updateGoPackagesInDir(servicePath)
	if err != nil {
		fmt.Printf("Ошибка выполнения: %v\n", err)
	} else {
		fmt.Println("Команда go get -u выполнена успешно")
	}
}
func updateGoPackagesInDir(path string) error {
	// Сохраняем текущую рабочую директорию
	originalDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("ошибка получения текущей рабочей директории: %v", err)
	}

	// Переходим в указанную директорию
	err = os.Chdir(path)
	if err != nil {
		return fmt.Errorf("ошибка изменения рабочей директории: %v", err)
	}

	// Выполняем команду "go get -u"
	cmd := exec.Command("go", "get", "-u")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("ошибка выполнения команды go get -u: %v", err)
	}

	// Возвращаемся в исходную рабочую директорию
	err = os.Chdir(originalDir)
	if err != nil {
		return fmt.Errorf("ошибка возврата в исходную рабочую директорию: %v", err)
	}

	return nil
}
