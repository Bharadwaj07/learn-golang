package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "add":
		cmdAdd(os.Args[2:])
	case "list", "ls":
		cmdList()
	case "done":
		cmdDone(os.Args[2:])
	case "delete", "rm":
		cmdDelete(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`tasks — a simple todo list

Usage:
  tasks add <title>    add a new task
  tasks list           show all tasks
  tasks done <id>      mark a task complete
  tasks delete <id>    remove a task`)
}

func cmdList() {
	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if len(tasks) == 0 {
		fmt.Println("No tasks yet. Try: tasks add Buy Milk")
		return
	}

	for _, t := range tasks {
		status := "[]"
		if t.Done {
			status = "[X]"
		}
		fmt.Printf("#%v %v  %v\n", t.ID, t.Title, status)
	}
}

func cmdAdd(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tasks add <title>")
		os.Exit(1)
	}
	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	task := Task{
		ID:        nextID(tasks),
		Title:     strings.Join(args, " "),
		Done:      false,
		CreatedAt: time.Now(),
	}

	tasks = append(tasks, task)

	if err := saveTasks(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Printf("Added #%d: %s\n", task.ID, task.Title)
}

func cmdDone(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tasks done <id>")
		os.Exit(1)
	}

	// convert the string argument "3" into an int
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", args[0])
		os.Exit(1)
	}

	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// find the task by ID and mark it done
	found := false
	for i, t := range tasks {
		if t.ID == id {
			tasks[i].Done = true // ← must use tasks[i], not t (see note below)
			found = true
			fmt.Printf("Marked #%d done: %s\n", t.ID, t.Title)
			break
		}
	}

	if !found {
		fmt.Fprintf(os.Stderr, "no task with id %d\n", id)
		os.Exit(1)
	}

	if err := saveTasks(tasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cmdDelete(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tasks delete <id>")
		os.Exit(1)
	}

	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid id: %s\n", args[0])
		os.Exit(1)
	}

	tasks, err := loadTasks()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// build a new slice with the target task filtered out
	newTasks := []Task{}
	found := false
	for _, t := range tasks {
		if t.ID == id {
			found = true
			fmt.Printf("Deleted #%d: %s\n", t.ID, t.Title)
			continue // skip this task — don't add it to newTasks
		}
		newTasks = append(newTasks, t)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "no task with id %d\n", id)
		os.Exit(1)
	}

	if err := saveTasks(newTasks); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
