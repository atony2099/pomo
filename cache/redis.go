package cache

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

type Cache struct {
	client *redis.Client
}

type SelectedTask struct {
	Name    string `json:"name"`
	TaskID  string `json:"task_id"`
	SubName string `json:"sub_name"`
	SubID   string `json:"sub_id"`
	Project string `json:"project"`
}

var redisClient *Cache

func NewClient(redisURL string) error {

	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
	})

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}
	redisClient = &Cache{
		client: client,
	}
	return nil
}

const SelectedTaskKey = "selected_task"
const PomodoroTimeKey = "pomodoro_time"

func SetSelectedTask(task SelectedTask) error {
	data, _ := json.Marshal(task)
	return redisClient.client.Set(SelectedTaskKey, data, 0).Err()
}

func GetSelectedTask() (SelectedTask, error) {
	data, err := redisClient.client.Get(SelectedTaskKey).Result()

	if err == redis.Nil {
		return SelectedTask{}, nil
	}

	if err != nil {
		return SelectedTask{}, err
	}

	var task SelectedTask
	err = json.Unmarshal([]byte(data), &task)
	return task, err
}
