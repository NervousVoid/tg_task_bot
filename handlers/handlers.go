package handlers

import (
	"_/models"
	"_/pkg"
	"_/utils"
	"fmt"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
)

func UnassignHandler(id string, user *tgbotapi.User, userChatID int64) (models.MessagesReply, error) {
	taskID, err := pkg.ValidateTaskID(id)
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID
	if err != nil {
		reply.MsgToUser = utils.CapitalizeFirstLetter(err.Error())
		return reply, fmt.Errorf(`error validating id "%s"`, id)
	}

	task := models.Tasks[taskID]
	if task.Assignee != nil && task.Assignee.ID != user.ID {
		reply.MsgToUser = utils.CapitalizeFirstLetter(utils.TaskNotAssignedError)
		return reply, nil
	}

	assignerChatID := pkg.UnassignTask(task)

	reply.MsgToAssigner = fmt.Sprintf(`Задача "%s" осталась без исполнителя`, task.Name)
	reply.AssignerChatID = assignerChatID

	reply.MsgToUser = "Принято"
	return reply, nil
}

func AssignHandler(id string, user *tgbotapi.User, userChatID int64) models.MessagesReply {
	taskID, err := pkg.ValidateTaskID(id)
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID
	if err != nil {
		reply.MsgToUser = utils.CapitalizeFirstLetter(err.Error())
		return reply
	}

	task := models.Tasks[taskID]

	prevAssigneeChatID := pkg.AssignTask(task, user, userChatID)
	reply.AssigneeChatID = prevAssigneeChatID

	notificationText := fmt.Sprintf(`Задача "%s" назначена на @%s`, task.Name, user)

	if prevAssigneeChatID != 0 {
		reply.MsgToPrevAssignee = notificationText
	} else if models.Tasks[taskID].AssignerChatID != userChatID {
		reply.MsgToAssigner = notificationText
		reply.AssignerChatID = task.AssignerChatID
	}

	reply.MsgToUser = fmt.Sprintf(`Задача "%s" назначена на вас`, models.Tasks[taskID].Name)
	return reply
}

func ResolveHandler(id string, user *tgbotapi.User, userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID

	taskID, err := pkg.ValidateTaskID(id)
	if err != nil {
		reply.MsgToUser = utils.CapitalizeFirstLetter(err.Error())
		return reply
	}

	task := models.Tasks[taskID]
	assignerChatID := pkg.ResolveTask(task)
	reply.AssignerChatID = assignerChatID

	if assignerChatID != userChatID {
		reply.MsgToAssigner = fmt.Sprintf(`Задача "%s" выполнена @%s`, task.Name, user)
	}

	reply.MsgToUser = fmt.Sprintf(`Задача "%s" выполнена`, task.Name)
	return reply
}

func HandleTasks(user *tgbotapi.User, userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID

	tasks := pkg.GetTasksSlice()
	resp := ""

	for pos, task := range tasks {
		resp += fmt.Sprintf("%d. %s by @%s\n", task.ID, task.Name, task.Assigner)

		switch {
		case task.Assignee == nil:
			resp += fmt.Sprintf("/assign_%d", task.ID)
		case task.Assignee.ID == user.ID:
			resp += fmt.Sprintf("assignee: я\n/unassign_%d /resolve_%d", task.ID, task.ID)
		case task.Assignee.ID != user.ID:
			resp += fmt.Sprintf("assignee: @%s", task.Assignee)
		}

		if pos != len(tasks)-1 {
			resp += "\n\n"
		}
	}

	if len(tasks) == 0 {
		resp = "Нет задач"
	}

	reply.MsgToUser = resp
	return reply
}

func HandleMy(user *tgbotapi.User, userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID

	resp := ""
	tasks := pkg.GetUserAssignations(user)
	if len(tasks) == 0 {
		resp = "На вас не назначено задач"
	} else {
		for pos, task := range tasks {
			resp += fmt.Sprintf("%d. %s by @%s\n/unassign_%d /resolve_%d",
				task.ID, task.Name, task.Assigner, task.ID, task.ID)

			if pos != len(tasks)-1 {
				resp += "\n\n"
			}
		}
	}
	reply.MsgToUser = resp
	return reply
}

func HandleOwner(user *tgbotapi.User, userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID

	resp := ""
	tasks := pkg.GetUserCreatedTasks(user)
	if len(tasks) == 0 {
		resp = "Нет активных задач"
	} else {
		for pos, task := range tasks {
			resp += fmt.Sprintf("%d. %s by @%s\n/assign_%d",
				task.ID, task.Name, task.Assigner, task.ID)

			if pos != len(tasks)-1 {
				resp += "\n\n"
			}
		}
	}

	reply.MsgToUser = resp
	return reply
}

func NewTaskHandler(taskName string, user *tgbotapi.User, userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.UserChatID = userChatID
	task := pkg.NewTask(taskName, user, userChatID)
	reply.MsgToUser = fmt.Sprintf(`Задача "%s" создана, id=%d`, task.Name, task.ID)
	return reply
}

func ReplyHandler(reply models.MessagesReply, bot *tgbotapi.BotAPI) error {
	if reply.MsgToUser != "" {
		msg := tgbotapi.NewMessage(reply.UserChatID, reply.MsgToUser)
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf(utils.CapitalizeFirstLetter(err.Error()))
		}
	}

	if reply.MsgToAssigner != "" {
		msg := tgbotapi.NewMessage(reply.AssignerChatID, reply.MsgToAssigner)
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf(utils.CapitalizeFirstLetter(err.Error()))
		}
	}

	if reply.MsgToPrevAssignee != "" {
		msg := tgbotapi.NewMessage(reply.AssigneeChatID, reply.MsgToPrevAssignee)
		_, err := bot.Send(msg)
		if err != nil {
			return fmt.Errorf(utils.CapitalizeFirstLetter(err.Error()))
		}
	}
	return nil
}

func StartHandler(userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.MsgToUser = utils.CommandsMessage
	reply.UserChatID = userChatID
	return reply
}

func NoSuchCommandHandler(userChatID int64) models.MessagesReply {
	reply := models.MessagesReply{}
	reply.MsgToUser = "Нет такой команды\nКоманды: /start"
	reply.UserChatID = userChatID
	return reply
}
