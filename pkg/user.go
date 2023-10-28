package pkg

import (
	"_/models"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

func GetUserAssignations(user *tgbotapi.User) []*models.Task {
	tasks := []*models.Task{}
	for _, task := range models.Tasks {
		if task.Assignee != nil && task.Assignee.ID == user.ID && task.IsActive {
			tasks = append(tasks, task)
		}
	}
	return tasks
}

func GetUserCreatedTasks(user *tgbotapi.User) []*models.Task {
	tasks := []*models.Task{}
	for _, task := range models.Tasks {
		if task.Assigner.ID == user.ID && task.IsActive {
			tasks = append(tasks, task)
		}
	}
	return tasks
}
