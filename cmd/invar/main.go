package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/user/invar/internal/app"
	"github.com/user/invar/internal/storage"
	"github.com/user/invar/internal/task"
)

func main() {
	var quickAdd string
	var quickNew bool
	flag.StringVar(&quickAdd, "n", "", "Quick add a new task")
	flag.BoolVar(&quickNew, "new", false, "Open input modal for quick task creation")
	flag.Parse()

	if quickAdd != "" {
		homeDir, _ := os.UserHomeDir()
		dataDir := fmt.Sprintf("%s/.local/share/invar/tasks", homeDir)
		store, err := storage.New(dataDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		t := task.New(quickAdd)
		if err := store.Save(t); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving task: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Task created:", quickAdd)
		return
	}

	m, err := app.New(quickNew)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
