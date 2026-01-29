package git

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Repo struct {
	path string
	repo *git.Repository
}

func Init(path string) (*Repo, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0755); err != nil {
			return nil, err
		}
	}

	gitDir := filepath.Join(path, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		repo, err := git.PlainInit(path, false)
		if err != nil {
			return nil, err
		}
		return &Repo{path: path, repo: repo}, nil
	}

	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}
	return &Repo{path: path, repo: repo}, nil
}

func (r *Repo) Commit(message string) error {
	w, err := r.repo.Worktree()
	if err != nil {
		return err
	}

	_, err = w.Add(".")
	if err != nil {
		return err
	}

	status, err := w.Status()
	if err != nil {
		return err
	}

	if status.IsClean() {
		return nil
	}

	_, err = w.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Invar",
			Email: "invar@localhost",
			When:  time.Now(),
		},
	})
	return err
}

func (r *Repo) Log() ([]string, error) {
	ref, err := r.repo.Head()
	if err != nil {
		return nil, err
	}

	iter, err := r.repo.Log(&git.LogOptions{From: ref.Hash()})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	var logs []string
	for {
		commit, err := iter.Next()
		if err != nil {
			break
		}
		logs = append(logs, fmt.Sprintf("%s %s %s", commit.Hash.String()[:7], commit.Author.When.Format("2006-01-02"), commit.Message))
	}
	return logs, nil
}
