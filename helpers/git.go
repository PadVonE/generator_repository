package helpers

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/config"
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

func CreateBranch(repoPath string, branchName string) error {
	// Открываем репозиторий
	r, err := git.PlainOpen(repoPath)
	if err != nil {
		return err
	}

	// Получаем HEAD-ссылку
	headRef, err := r.Head()
	if err != nil {
		return err
	}

	// Создаем новую ветку
	ref := plumbing.NewHashReference(plumbing.ReferenceName("refs/heads/"+branchName), headRef.Hash())

	// Сохраняем в конфиг
	err = r.Storer.SetReference(ref)
	if err != nil {
		return err
	}

	// Получаем Worktree
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	// Переключаемся на новую ветку
	err = w.Checkout(&git.CheckoutOptions{
		Branch: ref.Name(),
	})
	if err != nil {
		return err
	}

	// Сохраняем конфиг
	cfg, err := r.Config()
	if err != nil {
		return err
	}
	cfg.Branches[branchName] = &config.Branch{
		Name:   branchName,
		Remote: "origin",
		Merge:  plumbing.ReferenceName("refs/heads/" + branchName),
	}
	err = r.Storer.SetConfig(cfg)
	if err != nil {
		return err
	}

	return nil
}
