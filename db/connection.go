package db

import (
	"fmt"
	"strconv"
	"time"

	"github.com/atony2099/pomo/domain"
	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Task struct {
	Name         string
	Status       string
	ParentTaskID string
	ProjectName  string
	Duration     int64
	TaskID       string `gorm:"primaryKey"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

type TimeEntry struct {
	gorm.Model
	ID        string `gorm:"primaryKey"`
	TaskID    string
	StartTime time.Time
	EndTime   time.Time
	TaskName  string
}

type Activity struct {
	Date      string
	Tags      string
	StartTime string
	EndTime   string
	Do        string
}

type DB struct {
	db *gorm.DB
}

var dbs *DB

func Connect(dbDSN string) error {
	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return err
	}
	dbs = &DB{db}

	return nil
}

func GetTasks() ([]Task, error) {
	var tasks []Task
	result := dbs.db.Find(&tasks)
	return tasks, result.Error
}

// func (db *DB) InsertActivity(activity Activity) error {
// 	_, err := db.Exec("INSERT INTO daily_trackers (date, tags, start_time, end_time, do) VALUES (?, ?, ?, ?, ?)",
// 		activity.Date, activity.Tags, activity.StartTime, activity.EndTime, activity.Do)
// 	return err
// }

func InsertOrUpdateTasks(tasks []domain.TaskInfo, spaces []domain.Space) error {
	for _, taskInfo := range tasks {
		var spaceName string
		for _, space := range spaces {
			if space.ID == taskInfo.Space.ID {
				spaceName = space.Name
				break
			}
		}
		var task Task
		dbs.db.Where("task_id = ?", taskInfo.ID).First(&task)
		if task.TaskID != "" {
			task.Name = taskInfo.Name
			task.Status = taskInfo.Status.Status
			task.ParentTaskID = taskInfo.Parent
			task.ProjectName = spaceName
			err := dbs.db.Save(&task)
			if err.Error != nil {
				return fmt.Errorf("failed to update task %s: %v", task.TaskID, err.Error)
			}
		} else {
			task := Task{
				TaskID:       taskInfo.ID,
				Name:         taskInfo.Name,
				Status:       taskInfo.Status.Status,
				ParentTaskID: taskInfo.Parent,
				ProjectName:  spaceName,
			}
			err := dbs.db.Create(&task)
			if err.Error != nil {
				return fmt.Errorf("failed to insert task %s: %v", task.TaskID, err.Error)
			}
		}

	}
	return nil
}

func InsertOrUpdateEntries(entry []domain.TimeEntryInfo) error {

	for _, entry := range entry {

		startInt, err := strconv.ParseInt(entry.Start, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse start time: %v", err)
		}
		endInt, err := strconv.ParseInt(entry.End, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse end time: %v", err)
		}
		start := time.Unix(startInt/1000, 0)
		end := time.Unix(endInt/1000, 0)
		var timeEntry TimeEntry
		err = dbs.db.Where("id = ?", entry.ID).First(&timeEntry).Error
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("failed to fetch entry %s: %v", entry.ID, err)
		}

		if timeEntry.ID != "" {
			timeEntry.TaskID = entry.Task.ID
			timeEntry.StartTime = start
			timeEntry.EndTime = end
			timeEntry.TaskName = entry.Task.Name
			err := dbs.db.Model(&timeEntry).Updates(TimeEntry{TaskID: timeEntry.TaskID, StartTime: timeEntry.StartTime, EndTime: timeEntry.EndTime, TaskName: timeEntry.TaskName})
			if err.Error != nil {
				return fmt.Errorf("failed to update entry %s: %v", entry.ID, err.Error)
			}
		} else {
			timeEntry := TimeEntry{
				ID:        entry.ID,
				TaskID:    entry.Task.ID,
				StartTime: start,
				EndTime:   end,
				TaskName:  entry.Task.Name,
			}
			err := dbs.db.Create(&timeEntry)
			if err.Error != nil {
				return fmt.Errorf("failed to insert entry %s: %v", entry.ID, err.Error)
			}
		}
	}
	return nil
}

func SumTaskDurationsAndUpdateTask() error {

	// USE GORM
	var taskDurations []*domain.TaskDuration
	err := dbs.db.Table("time_entries").Select("task_id, SUM(TIMESTAMPDIFF(SECOND, start_time, end_time)) as duration").Group("task_id").Scan(&taskDurations).Error

	if err != nil {
		return fmt.Errorf("failed to sum task durations: %v", err)
	}

	for _, taskDuration := range taskDurations {
		// check if task have child tasks
		var taskIDs []string
		dbs.db.Table("tasks").Where("parent_task_id = ?", taskDuration.TaskID).Pluck("task_id", &taskIDs)
		if len(taskIDs) > 0 {
			// sum child tasks durations
			var totalDuration int64
			for _, taskID := range taskIDs {
				for _, taskDuration := range taskDurations {
					if taskID == taskDuration.TaskID {
						totalDuration += taskDuration.Duration
					}
				}
			}
			taskDuration.Duration += totalDuration
		}
	}

	for _, taskDuration := range taskDurations {
		dbs.db.Table("tasks").Where("task_id = ?", taskDuration.TaskID).Update("duration", taskDuration.Duration)
	}

	return nil
}

func SelectTimeEntry(day string) ([]domain.TimeEntry, error) {
	var tasks []domain.TimeEntry

	err := dbs.db.Where("date(start_time) = ? or date(end_time) = ?", day, day).Find(&tasks).Error
	return tasks, err
}

func SaveTimeEntry(entry domain.TimeEntry) error {
	err := dbs.db.Create(&entry).Error
	return err
}

// CREATE TABLE `time_entries` (
//
//	`id` varchar(255) NOT NULL,
//	`task_id` varchar(255) NOT NULL,
//	`task_name` varchar(255) NOT NULL,
//	`start_time` datetime NOT NULL,
//	`end_time` datetime NOT NULL,
//	`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
//	`updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//	`deleted_at` datetime DEFAULT NULL,
//	PRIMARY KEY (`id`)
//
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
func SelectTotalDurationGroupByTask() ([]domain.TaskSummary, error) {
	var taskDurations []domain.TaskSummary

	err := dbs.db.Table("time_entries").Select("task_name, SUM(TIMESTAMPDIFF(SECOND, start_time, end_time)) as duration").Group("task_name").Scan(&taskDurations).Error

	// the origin sql is :
	// SELECT task_name, SUM(TIMESTAMPDIFF(SECOND, start_time, end_time)) as duration FROM time_entries GROUP BY task_name
	return taskDurations, err

}

func CreateDailyTracker(tracker domain.DailyTracker) error {

	// check if the tracker.start_time already exists
	var count int64
	dbs.db.Model(&domain.DailyTracker{}).Where("start_time = ?", tracker.StartTime).Count(&count)
	if count > 0 {
		return nil
	}

	err := dbs.db.Create(&tracker).Error
	return err
}

func SelectDailyTracker(date string) ([]domain.DailyTracker, error) {
	var trackers []domain.DailyTracker

	// selct end_time  is date ,
	// if end_time is null, select start_time is date

	err := dbs.db.Where("date(start_time) = ? or date(end_time) = ?", date, date).Order("start_time").Find(&trackers).Error
	return trackers, err
}

// func SelectTaskForSomeDay(day string) ([]domain.TimeEntry, error) {
// 	var entries []domain.TimeEntry

// 	err := dbs.db.Where("date(start_time) = ?", day).Find(&entries).Error

// 	return entries, err

// }

func GetAllActivityName() ([]string, error) {

	var activities []string

	err := dbs.db.Table("daily_trackers").Select("activity").Group("activity").Scan(&activities).Error

	return activities, err

}

func UpdateDailyTracker(tracker domain.DailyTracker) error {

	// where start_time = tracker.start_time and activity = tracker.activity
	err := dbs.db.Model(&domain.DailyTracker{}).Where("start_time = ? and activity = ?", tracker.StartTime, tracker.Activity).Updates(&tracker).Error
	return err
}
