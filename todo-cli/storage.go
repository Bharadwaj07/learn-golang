package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func storePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "tasks.json"
	}
	return filepath.Join(home, ".tasks.json")
}

func loadTasks() ([]Task, error) {
	data, err := os.ReadFile(storePath())

	if os.IsNotExist(err) {
		return []Task{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	var tasks []Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("parsing tasks: %w", err)
	}

	return tasks, nil
}

func saveTasks(tasks []Task) error {
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding tasks: %w", err)
	}
	return os.WriteFile(storePath(), data, 0644)
}

func nextID(tasks []Task) int {
	max := 0
	for _, t := range tasks {
		if t.ID > max {
			max = t.ID
		}
	}
	return max + 1
}
