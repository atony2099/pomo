package task

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/atony2099/pomo/db"
	"github.com/atony2099/pomo/domain"
)

const baseURL = "https://api.clickup.com/api/v2"

type ClickUpClient struct {
	apiToken string
	teamID   string
	client   *http.Client
}

// Constructors
func NewClickUpClient(apiToken, teamID string) *ClickUpClient {
	return &ClickUpClient{
		apiToken: apiToken,
		teamID:   teamID,
		client:   &http.Client{},
	}
}

// API call helper
func (c *ClickUpClient) makeRequest(endpoint string) ([]byte, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", baseURL, endpoint), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", c.apiToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var apiErr domain.APIError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err != nil {
			return nil, fmt.Errorf("failed to decode API error: %v", err)
		}
		return nil, fmt.Errorf("API error: %s - %s", apiErr.Error, apiErr.Code)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return data, nil
}

func (c *ClickUpClient) GetSpaces() ([]domain.Space, error) {
	data, err := c.makeRequest(fmt.Sprintf("/team/%s/space", c.teamID))
	if err != nil {
		return nil, err
	}

	var response domain.SpaceResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spaces: %v", err)
	}

	return response.Spaces, nil
}

// GetLists retrieves lists for a space
func (c *ClickUpClient) GetLists(spaceID string) ([]domain.List, error) {
	data, err := c.makeRequest(fmt.Sprintf("/space/%s/list", spaceID))
	if err != nil {
		return nil, err
	}

	var response domain.ListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal lists: %v", err)
	}

	return response.Lists, nil
}

// GetTasks retrieves tasks for a list
func (c *ClickUpClient) GetTasks(listID string) ([]domain.TaskInfo, error) {
	data, err := c.makeRequest(fmt.Sprintf("/list/%s/task?subtasks=true", listID))
	if err != nil {
		return nil, err
	}

	var taskResp domain.TaskResponse
	if err := json.Unmarshal(data, &taskResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tasks: %v", err)
	}

	return taskResp.Tasks, nil
}

// InsertTasks inserts tasks into the database

func SyncData(apiToken, teamID string) {
	client := NewClickUpClient(apiToken, teamID)

	spaces, err := client.GetSpaces()
	if err != nil {
		log.Fatalf("Failed to fetch spaces: %v", err)
	}
	for _, space := range spaces {
		lists, err := client.GetLists(space.ID)
		if err != nil {
			log.Fatalf("Failed to fetch lists for space %s: %v", space.ID, err)
			continue
		}

		for _, list := range lists {
			tasks, err := client.GetTasks(list.ID)
			if err != nil {
				log.Fatalf("Failed to fetch tasks for list %s: %v", list.ID, err)
				continue
			}

			if err := db.InsertOrUpdateTasks(tasks, spaces); err != nil {
				log.Printf("Failed to insert tasks for list %s: %v", list.ID, err)
			}

		}
	}

	// entries, err := client.getEntries()
	// if err != nil {
	// 	log.Fatalf("Failed to fetch time entries: %v", err)
	// }

	// if err := db.InsertOrUpdateEntries(entries); err != nil {
	// 	log.Fatalf("Failed to insert time entries: %v", err)
	// }

	// if err := db.SumTaskDurationsAndUpdateTask(); err != nil {
	// 	log.Fatalf("Failed to sum task durations: %v", err)
	// }

}

func (c *ClickUpClient) getEntries() ([]domain.TimeEntryInfo, error) {
	data, err := c.makeRequest(fmt.Sprintf("/team/%s/time_entries", c.teamID))
	if err != nil {
		return nil, err
	}

	var response domain.TimeEntryResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spaces: %v", err)
	}

	return response.Data, nil
}

// CREATE TABLE time_entries (
//   id int NOT NULL AUTO_INCREMENT,
//   task_id varchar(255) NOT NULL,
//   start_time datetime NOT NULL,
//   end_time datetime NOT NULL,
//   created_at datetime NOT NULL default current_timestamp,
//   updated_at datetime NOT NULL default current_timestamp on update current_timestamp,
//   deleted_at datetime,
//   PRIMARY KEY (id)
// );
