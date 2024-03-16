package task

import (
	"fmt"

	"strconv"
	"strings"

	"github.com/atony2099/pomo/cache"
	"github.com/atony2099/pomo/db"
)

// DisplayTasks prints the list of tasks and returns them.
func DisplayTasks() ([]db.Task, error) {
	tasks, err := db.GetTasks()
	if err != nil {
		return nil, fmt.Errorf("error retrieving tasks: %w", err)
	}

	var mainTasks []db.Task
	for _, task := range tasks {
		if task.ParentTaskID == "" {
			mainTasks = append(mainTasks, task)
			fmt.Printf("%d. %s (%s):\n", len(mainTasks), task.Name, task.ProjectName)
			displaySubtasks(task.TaskID, tasks)
		}
	}
	return tasks, nil
}

// displaySubtasks helps DisplayTasks by printing subtasks of a given main task.
func displaySubtasks(mainTaskID string, tasks []db.Task) {
	subtaskCount := 0
	for _, task := range tasks {
		if task.ParentTaskID == mainTaskID {
			subtaskCount++
			fmt.Printf(" [%d]. %s\n", subtaskCount, task.Name)
		}
	}
}

// SetPomodoroConfig sets the configuration for Pomodoro based on user input.
func SetPomodoroConfig() {
	tasks, err := DisplayTasks()
	if err != nil {
		fmt.Printf("Error displaying tasks: %v\n", err)
		return
	}

	selectedTask, err := getSelectedTask(tasks)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Printf("Selected task: %s %s, Project: %s\n", selectedTask.Name, selectedTask.SubName, selectedTask.Project)
	if err := cache.SetSelectedTask(selectedTask); err != nil {
		fmt.Printf("Error setting selected task: %v\n", err)
	}
}

// getSelectedTask prompts the user to select a task and returns the selected task.
func getSelectedTask(tasks []db.Task) (cache.SelectedTask, error) {
	currentTask, err := cache.GetSelectedTask()
	if err != nil {
		return cache.SelectedTask{}, fmt.Errorf("error getting selected task: %w", err)
	}
	fmt.Printf("Current task: %s %s\nEnter the task number (e.g., 1.1): ", currentTask.Name, currentTask.SubName)

	var input string
	if _, err := fmt.Scanln(&input); err != nil {
		return cache.SelectedTask{}, fmt.Errorf("failed to read input: %w", err)
	}
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("Using previous task")
		return currentTask, nil
	}

	mainNum, subNum, err := parseTaskInput(input)
	if err != nil {
		return cache.SelectedTask{}, err
	}

	return selectTask(tasks, mainNum, subNum)
}

// parseTaskInput parses the user input into main and sub task numbers.
func parseTaskInput(input string) (int, int, error) {
	parts := strings.Split(input, ".")
	mainNum, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, fmt.Errorf("invalid main task number: %w", err)
	}
	subNum := 0
	if len(parts) == 2 {
		subNum, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid sub task number: %w", err)
		}
	}
	return mainNum, subNum, nil
}

// selectTask finds and returns the task based on main and sub task numbers.
func selectTask(tasks []db.Task, mainNum, subNum int) (cache.SelectedTask, error) {
	var mainTasks, selectedSubTasks []db.Task
	for _, task := range tasks {
		if task.ParentTaskID == "" {
			mainTasks = append(mainTasks, task)
		}
	}

	if mainNum <= 0 || mainNum > len(mainTasks) {
		return cache.SelectedTask{}, fmt.Errorf("main task number out of range")
	}

	selectedMainTask := mainTasks[mainNum-1]
	for _, task := range tasks {
		if task.ParentTaskID == selectedMainTask.TaskID {
			selectedSubTasks = append(selectedSubTasks, task)
		}
	}

	if subNum <= 0 {
		return createSelectedTask(selectedMainTask, nil), nil
	}

	if subNum > len(selectedSubTasks) {
		return cache.SelectedTask{}, fmt.Errorf("sub task number out of range")
	}

	selectedSubTask := selectedSubTasks[subNum-1]
	return createSelectedTask(selectedMainTask, &selectedSubTask), nil

}

// createSelectedTask creates a cache.SelectedTask from the db.Task(s) provided.
func createSelectedTask(mainTask db.Task, subTask *db.Task) cache.SelectedTask {
	selectedTask := cache.SelectedTask{
		Name:    mainTask.Name,
		TaskID:  mainTask.TaskID,
		Project: mainTask.ProjectName,
	}

	if subTask != nil {
		selectedTask.SubName = subTask.Name
		selectedTask.SubID = subTask.TaskID
		if subTask.ProjectName != "" {
			selectedTask.Project = subTask.ProjectName
		}
	}

	return selectedTask
}
