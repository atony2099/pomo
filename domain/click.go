package domain

import (
	"time"

	"gorm.io/gorm"
)

// APIError represents a ClickUp API error response
type APIError struct {
	Error string `json:"err"`
	Code  string `json:"ECODE"`
}

// Space represents a space in ClickUp
type Space struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SpaceResponse struct {
	Spaces []Space `json:"spaces"`
}

type ListResponse struct {
	Lists []List `json:"lists"`
}

// List represents a list in ClickUp
type List struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Task represents a task in ClickUp
type TaskInfo struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status struct {
		Status string `json:"status"`
	} `json:"status"`
	Parent string `json:"parent"`
	Space  struct {
		ID string `json:"id"`
	} `json:"space"`
}

type TaskResponse struct {
	Tasks []TaskInfo `json:"tasks"`
}

type TimeEntryInfo struct {
	ID    string   `json:"id"`
	Task  TaskInfo `json:"task"`
	Start string   `json:"start"`
	End   string   `json:"end"`
}

type TimeEntryResponse struct {
	Data []TimeEntryInfo `json:"data"`
}

type TaskDuration struct {
	TaskID   string
	Duration int64
}

type TaskSummary struct {
	TaskName string
	Duration int64
}

// CREATE TABLE `time_entries` (
//   `id` varchar(255) NOT NULL,
//   `task_id` varchar(255) NOT NULL,
//   `task_name` varchar(255) NOT NULL,
//   `start_time` datetime NOT NULL,
//   `end_time` datetime NOT NULL,
//   `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
//   `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
//   `deleted_at` datetime DEFAULT NULL,
//   PRIMARY KEY (`id`)
// ) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

type TimeEntry struct {
	gorm.Model
	TaskID    string
	TaskName  string
	StartTime time.Time
	EndTime   time.Time
}

type DailyTracker struct {
	ID        uint `gorm:"primaryKey"`
	Activity  string
	StartTime time.Time
	EndTime   time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}
