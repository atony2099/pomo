package task

import (
	"context"
	"fmt"

	"time"

	"github.com/atony2099/pomo/audio"
	"github.com/atony2099/pomo/cache"
	"github.com/atony2099/pomo/db"
	"github.com/atony2099/pomo/domain"
	"github.com/atony2099/pomo/ui"
	"github.com/nsf/termbox-go"
)

type TaskHandler struct {
	pomodoroDuration time.Duration
	stopInFirst      time.Duration
	authKey          string
	breakDuration    time.Duration
	teamID           string
}

func NewTaskHandler(authKey, teamID string, pomodortime, invalidtime, breaktime int) *TaskHandler {
	return &TaskHandler{
		pomodoroDuration: time.Duration(pomodortime) * time.Minute,
		stopInFirst:      time.Duration(invalidtime) * time.Second,
		authKey:          authKey,
		breakDuration:    time.Duration(breaktime) * time.Minute,
		teamID:           teamID,
	}
}

func listenForExit(exitChan chan bool) {
	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeySpace {
				exitChan <- true
			}
		}
	}
}

func (h *TaskHandler) RunPomodoro() {
	startTime := time.Now()
	exitChan := make(chan bool, 1)
	go listenForExit(exitChan)

	timerTick := time.NewTicker(1 * time.Second)
	defer timerTick.Stop()

	for {
		select {
		case <-exitChan:
			h.finishPomodoro(startTime, time.Now(), audio.Interrupt, exitChan)
			return
		case <-timerTick.C:
			elapsed := time.Since(startTime)
			if elapsed > h.pomodoroDuration {
				h.finishPomodoro(startTime, time.Now(), audio.Finish, exitChan)
				return
			}
			ui.DrawCountdownFull(h.pomodoroDuration, elapsed)
		}
	}
}

func (task *TaskHandler) finishPomodoro(start, end time.Time, soundType audio.SoundType, exitChan chan bool) {
	termbox.Close()

	if end.Sub(start) <= task.stopInFirst {
		fmt.Printf("pomo duration: %ds less than %v seconds, ignore it\n", end.Sub(start)/time.Second, task.stopInFirst)
		return
	}

	// excute the sync task

	err := task.saveTimeEntry(context.Background(), start, end)
	if err != nil {
		fmt.Printf("error posting data: %v\n", err)
		return
	}

	audio.PlaySound(soundType)

	ui.ClearScreen()
	if soundType == audio.Finish {
		go func() {
			// SyncData(task.authKey, task.teamID)
			Complete(0)
		}()
		task.runBreakTimer(exitChan)
	} else {
		// SyncData(task.authKey, task.teamID)
		Complete(0)
	}

}

// runBreakTimer manages the break period after a pomodoro session.
func (h *TaskHandler) runBreakTimer(exitChan chan bool) {
	startTime := time.Now()
	breakTicker := time.NewTicker(1 * time.Second)
	defer breakTicker.Stop()

	for {
		select {
		case <-exitChan:
			fmt.Println("Break stopped early.")
			audio.PlaySound(audio.Breaks)
			return
		case <-breakTicker.C:
			elapsed := time.Since(startTime)
			ui.DrawProgressBar(elapsed.Seconds(), h.breakDuration.Seconds())
			if elapsed > h.breakDuration {
				audio.PlaySound(audio.Breaks)
				return
			}
		}
	}
}

func (h *TaskHandler) saveTimeEntry(ctx context.Context, start, end time.Time) error {

	// get the selected task
	task, err := cache.GetSelectedTask()
	if err != nil {
		return fmt.Errorf("error getting selected task: %v", err)
	}

	taskID := task.TaskID
	if task.SubID != "" {
		taskID = task.SubID
	}

	// geneternage unique id
	id := fmt.Sprintf("%s-%d", taskID, time.Now().UnixNano())
	time := domain.TimeEntry{
		ID:        id,
		TaskID:    taskID,
		StartTime: start,
		EndTime:   end,
	}

	err = db.SaveTimeEntry(time)

	if err != nil {
		return fmt.Errorf("error saving time entry: %v", err)
	}

	return nil

}

// func (h *TaskHandler) saveTimeEntry(ctx context.Context, start, end time.Time) error {
// 	url := fmt.Sprintf("https://api.clickup.com/api/v2/team/%s/time_entries", h.teamID)
// 	task, err := cache.GetSelectedTask()
// 	if err != nil {
// 		return fmt.Errorf("error getting selected task: %v", err)
// 	}

// 	taskID := task.TaskID
// 	if task.SubID != "" {
// 		taskID = task.SubID
// 	}

// 	postData := map[string]interface{}{
// 		"start":    start.Unix() * 1000,
// 		"duration": int(end.Sub(start).Milliseconds()),
// 		"tid":      taskID,
// 	}
// 	data, err := json.Marshal(postData)
// 	if err != nil {
// 		return err
// 	}

// 	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
// 	if err != nil {
// 		return fmt.Errorf("error creating request: %v", err)
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Authorization", h.authKey)

// 	client := &http.Client{}
// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return fmt.Errorf("error sending request: %v", err)
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		body, _ := io.ReadAll(resp.Body)
// 		return fmt.Errorf("response error: %s - %s", resp.Status, body)
// 	}

// 	return nil
// }
