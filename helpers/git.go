package helpers

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"os/exec"
	"path/filepath"
)

func GitCheckoutBranch(repoPath string, branch string) error {
	cmd := exec.Command("git", "checkout", branch)
	cmd.Dir = filepath.Clean(repoPath)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ошибка выполнения команды: %v\nOutput: %s", err, output)
	}
	return nil
}

// GitRepoInfo содержит информацию о репозитории
type GitRepoInfo struct {
	CurrentBranch string
	ChangedFiles  int
	BranchList    []string
}

// GetGitRepoInfo возвращает информацию о репозитории Git
func GetGitRepoInfo(repoPath string) (getRepo *GitRepoInfo, err error) {
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return getRepo, fmt.Errorf("Ошибка открытия репозитория: %w", err)
	}

	// Получаем ветку, на которой мы сейчас находимся
	head, err := r.Head()
	if err != nil {
		return getRepo, fmt.Errorf("Ошибка получения информации о ветке: %w", err)
	}

	// Получаем информацию о статусе репозитория
	w, err := r.Worktree()
	if err != nil {
		return getRepo, fmt.Errorf("Ошибка получения рабочей области: %w", err)
	}

	status, err := w.Status()
	if err != nil {
		return getRepo, fmt.Errorf("Ошибка получения статуса: %w", err)
	}

	// Количество изменённых файлов
	var modifiedFiles int
	for _, s := range status {
		if s.Staging != git.Unmodified || s.Worktree != git.Unmodified {
			modifiedFiles++
		}
	}

	// получаем список ссылок (refs)
	refs, err := r.References()
	if err != nil {
		return nil, err
	}

	branchList := []string{}
	// итерируем по ссылкам и добавляем имена веток в список
	refs.ForEach(func(ref *plumbing.Reference) error {
		if ref.Name().IsBranch() {
			branchList = append(branchList, ref.Name().Short())
		}
		return nil
	})

	return &GitRepoInfo{
		CurrentBranch: head.Name().Short(),
		ChangedFiles:  modifiedFiles,
		BranchList:    branchList,
	}, nil
}
