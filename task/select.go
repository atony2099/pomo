package task

import (
	"fmt"

	"time"

	"github.com/atony2099/pomo/db"
)

func SelectTask(offset int, isTotal bool) {

	if isTotal {
		// do something
		list, err := db.SelectTotalDurationGroupByTask()
		if err != nil {
			fmt.Printf("Error selecting total duration group by task: %v\n", err)
			return
		}
		for _, l := range list {
			duration := time.Duration(l.Duration) * time.Second
			fmt.Printf("%-10s: %v\n", l.TaskName, duration)
		}
		return
	}

	fmt.Printf("Selecting task from %d days ago\n", offset)

	day := time.Now().AddDate(0, 0, -offset).Format("2006-01-02")
	// fmt.Printf("Start time: %s\n", startTime)
	list, err := db.SelectTimeEntry(day)
	if err != nil {

		fmt.Printf("Error selecting task from day: %v\n", err)
		return
	}

	// ouput ： task_name, start_time, end_time, duration
	for _, l := range list {
		duration := l.EndTime.Sub(l.StartTime)
		fmt.Printf("%-10s: %s - %s, %v\n", l.TaskName, l.StartTime.Format("2006-01-02 15:04:05"), l.EndTime.Format("2006-01-02 15:04:05"), duration)
	}
	// get total duration for every day

	var maps = make(map[string]time.Duration)

	for _, l := range list {

		// get the date of the start time
		date := l.StartTime.Format("2006-01-02")

		if _, ok := maps[date]; !ok {
			maps[date] = l.EndTime.Sub(l.StartTime)
		} else {
			maps[date] += l.EndTime.Sub(l.StartTime)
		}
	}

	for k, v := range maps {
		fmt.Printf("%s: %v\n", k, v)
	}
}
