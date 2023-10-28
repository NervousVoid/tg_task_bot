package models

import tgbotapi "github.com/skinass/telegram-bot-api/v5"

type Task struct {
	ID             int
	Assignee       *tgbotapi.User
	AssigneeChatID int64
	Assigner       *tgbotapi.User
	AssignerChatID int64
	Name           string
	IsActive       bool
}

var Tasks = make(map[int]*Task)
