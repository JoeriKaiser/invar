package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/user/invar/internal/git"
	"github.com/user/invar/internal/task"
)

type Store struct {
	dataDir string
	repo    *git.Repo
}

func New(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	repo, err := git.Init(dataDir)
	if err != nil {
		return nil, err
	}

	return &Store{dataDir: dataDir, repo: repo}, nil
}

func (s *Store) Save(t *task.Task) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	filename := filepath.Join(s.dataDir, t.ID+".json")
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return err
	}

	return s.repo.Commit(fmt.Sprintf("Update task: %s", t.ID[:8]))
}

func (s *Store) Load(id string) (*task.Task, error) {
	filename := filepath.Join(s.dataDir, id+".json")
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var t task.Task
	if err := json.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Store) Delete(id string) error {
	filename := filepath.Join(s.dataDir, id+".json")
	if err := os.Remove(filename); err != nil {
		return err
	}
	return s.repo.Commit(fmt.Sprintf("Delete task: %s", id[:8]))
}

func (s *Store) List(archived bool) ([]*task.Task, error) {
	entries, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, err
	}

	var tasks []*task.Task
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		id := entry.Name()[:len(entry.Name())-5]
		t, err := s.Load(id)
		if err != nil {
			continue
		}
		if t.Archived == archived {
			tasks = append(tasks, t)
		}
	}
	return tasks, nil
}

func (s *Store) DataDir() string {
	return s.dataDir
}

func (s *Store) Log() ([]string, error) {
	return s.repo.Log()
}
