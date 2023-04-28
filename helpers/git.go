package helpers

import (
	"fmt"
	"os/exec"
	"path/filepath"
)

func GitCheckoutDev(repoPath string) error {
	cmd := exec.Command("git", "checkout", "dev")
	cmd.Dir = filepath.Clean(repoPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка выполнения команды: %v\nOutput: %s", err, output)
	}
	return nil
}
