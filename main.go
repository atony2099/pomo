package main

import (
	"flag"
	"log"

	"github.com/atony2099/pomo/cache"
	"github.com/atony2099/pomo/config"
	"github.com/atony2099/pomo/db"
	"github.com/atony2099/pomo/task"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nsf/termbox-go"
)

// ... other imports
func main() {

	var setFlag = flag.Bool("set", false, "set pomodoro config")
	var taskFlag = flag.Bool("sync", false, "get task list")
	// select specify day

	var total = flag.Bool("total", false, "total duration")

	var completeFlag = flag.Int("complete", -1, "complete task")

	var logFlg = flag.Int("log", -1, "select the activity log")

	// set start flag
	var startFlag = flag.Bool("start", false, "start activity")

	flag.Parse()

	config := config.LoadConfig()
	//
	err := cache.NewClient(config.RedisURL)
	if err != nil {
		log.Fatalf("Error initializing cache: %v", err)
	}

	//
	err = db.Connect(config.DBDSN)
	if err != nil {
		log.Fatalf("Error initializing database: %v", err)
	}

	if *setFlag {
		task.SetPomodoroConfig()
		return
	}
	if *taskFlag {
		task.SyncData(config.AuthKey, config.TeamID)
		return
	}

	if *logFlg >= 0 {
		task.GetActivities(*logFlg)
		return
	}

	if *total {
		task.SelectTask(0, true)
		return
	}

	if *completeFlag >= 0 {
		task.Complete(*completeFlag)
		return
	}

	if *startFlag {
		task.CrateActiveStart()
		return
	}

	err = termbox.Init()
	if err != nil {
		log.Fatalf("Error initializing termbox: %v", err)
	}
	defer termbox.Close()
	task := task.NewTaskHandler(config.AuthKey, config.TeamID, config.PomodoroTime, config.StopInFirst, config.BreakTime)

	task.RunPomodoro()

}
