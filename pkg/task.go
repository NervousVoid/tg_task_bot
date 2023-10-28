package pkg

import (
	"_/models"
	"_/utils"
	"fmt"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"sort"
	"strconv"
)

func GetTasksSlice() []*models.Task {
	tasksSlice := []*models.Task{}

	for _, task := range models.Tasks {
		if task.IsActive {
			tasksSlice = append(tasksSlice, task)
		}
	}

	sort.Slice(tasksSlice, func(i, j int) bool {
		return tasksSlice[i].ID < tasksSlice[j].ID
	})

	return tasksSlice
}

func NewTask(name string, user *tgbotapi.User, chatID int64) *models.Task {
	task := &models.Task{
		ID:             len(models.Tasks) + 1,
		Assigner:       user,
		AssignerChatID: chatID,
		Name:           name,
		IsActive:       true,
	}

	models.Tasks[task.ID] = task
	return task
}

func ValidateTaskID(id string) (int, error) {
	taskID, err := strconv.Atoi(id)
	if err != nil {
		return -1, fmt.Errorf(utils.InvalidIDError)
	}
	if task, ok := models.Tasks[taskID]; !ok || !task.IsActive {

		return -1, fmt.Errorf(utils.IDNotFoundError)
	}
	return taskID, nil
}

func AssignTask(task *models.Task, user *tgbotapi.User, userChatID int64) int64 {
	prevAssigneeChatID := task.AssigneeChatID
	task.Assignee = user
	task.AssigneeChatID = userChatID
	return prevAssigneeChatID
}

func UnassignTask(task *models.Task) int64 {
	task.Assignee = nil
	task.AssigneeChatID = 0
	return task.AssignerChatID
}

func ResolveTask(task *models.Task) int64 {
	task.IsActive = false
	return task.AssignerChatID
}
