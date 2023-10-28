package main

import (
	"_/handlers"
	"_/models"
	"context"
	"fmt"
	"github.com/joho/godotenv"
	tgbotapi "github.com/skinass/telegram-bot-api/v5"
	"log"
	"net/http"
	"os"
)

const (
	newCmd      = "/new "
	assignCmd   = "/assign_"
	unassignCmd = "/unassign_"
	resolveCmd  = "/resolve_"

	port = "8080"
)

var (
	newCmdLen      = len(newCmd)
	assignCmdLen   = len(assignCmd)
	unassignCmdLen = len(unassignCmd)
	resolveCmdLen  = len(resolveCmd)
)

func startBot(botToken, webhookURL string) (*tgbotapi.BotAPI, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, fmt.Errorf("NewBotAPI failed %s", err)
	}

	bot.Debug = true
	fmt.Printf("Authorized on account %s", bot.Self.UserName)

	wh, err := tgbotapi.NewWebhook(webhookURL)
	if err != nil {
		return nil, fmt.Errorf("NewWebhook failed: %s", err)
	}

	_, err = bot.Request(wh)
	if err != nil {
		return nil, fmt.Errorf("SetWebhook failed: %s", err)
	}

	go func() {
		log.Fatalln("http err:", http.ListenAndServe(":"+port, nil))
	}()
	fmt.Println("start listen: " + port)
	return bot, nil
}

func StartTaskBot(ctx context.Context, bt, wh string) error {
	var reply models.MessagesReply
	var bot *tgbotapi.BotAPI

	bot, err := startBot(bt, wh)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	updates := bot.ListenForWebhook("/")
	for upd := range updates {
		if upd.Message == nil {
			continue
		}
		log.Printf("upd: %#v\n", upd)

		msgText := upd.Message.Text
		switch {
		case msgText == "/tasks":
			reply = handlers.HandleTasks(upd.Message.From, upd.Message.Chat.ID)
		case len(msgText) > newCmdLen && msgText[0:newCmdLen] == "/new ":
			reply = handlers.NewTaskHandler(msgText[newCmdLen:], upd.Message.From, upd.Message.Chat.ID)
		case len(msgText) > assignCmdLen && msgText[0:assignCmdLen] == "/assign_":
			reply = handlers.AssignHandler(msgText[assignCmdLen:], upd.Message.From, upd.Message.Chat.ID)
		case len(msgText) > unassignCmdLen && msgText[0:unassignCmdLen] == "/unassign_":
			reply, err = handlers.UnassignHandler(msgText[unassignCmdLen:], upd.Message.From, upd.Message.Chat.ID)
			if err != nil {
				return err
			}
		case len(msgText) > resolveCmdLen && msgText[0:resolveCmdLen] == "/resolve_":
			reply = handlers.ResolveHandler(msgText[resolveCmdLen:], upd.Message.From, upd.Message.Chat.ID)
		case msgText == "/my":
			reply = handlers.HandleMy(upd.Message.From, upd.Message.Chat.ID)
		case msgText == "/owner":
			reply = handlers.HandleOwner(upd.Message.From, upd.Message.Chat.ID)
		case msgText == "/start":
			reply = handlers.StartHandler(upd.Message.Chat.ID)
		default:
			reply = handlers.NoSuchCommandHandler(upd.Message.Chat.ID)
		}
		err = handlers.ReplyHandler(reply, bot)
		if err != nil {
			return err
		}
	}
	return nil
}
func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	wh := os.Getenv("WEBHOOK_URL")
	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Fatal("BOT_TOKEN not found")
	}
	if wh == "" {
		log.Fatal("WEBHOOK_URL not found")
	}
	fmt.Println(wh)
	err := StartTaskBot(context.Background(), token, wh)
	if err != nil {
		panic(err)
	}
}
