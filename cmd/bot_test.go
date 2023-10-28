package main

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync/atomic"

	// "encoding/json"
	"fmt"

	tgbotapi "github.com/skinass/telegram-bot-api/v5"

	// "io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

var (
	BotToken   = "testToken"
	WebhookURL = ""
	client     = &http.Client{Timeout: time.Second}
)

func init() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Fatal("Error loading .env file")
	}
	WebhookURL = os.Getenv("WEBHOOK_URL")

	if WebhookURL == "" {
		log.Fatal("WEBHOOK_URL not found")
	}

}

// TDS is Telegram Dummy Server
type TDS struct {
	*sync.Mutex
	Answers map[int64]string
}

func NewTDS() *TDS {
	return &TDS{
		Mutex:   &sync.Mutex{},
		Answers: make(map[int64]string),
	}
}

func (srv *TDS) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux := http.NewServeMux()
	mux.HandleFunc("/getMe", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`{"ok":true,"result":{"id":` +
			strconv.Itoa(BotChatID) +
			`,"is_bot":true,"first_name":"game_test_bot","username":"game_test_bot"}}`))
	})
	mux.HandleFunc("/setWebhook", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		w.Write([]byte(`{"ok":true,"result":true,"description":"Webhook was set"}`))
	})
	mux.HandleFunc("/sendMessage", func(w http.ResponseWriter, r *http.Request) {
		//nolint:errcheck
		chatID, _ := strconv.ParseInt(r.FormValue("chat_id"), 10, 64)
		text := r.FormValue("text")
		srv.Lock()
		srv.Answers[chatID] = text
		srv.Unlock()

		//nolint:errcheck
		w.Write([]byte(`{"ok":true, "result":{"MessageID": 0}}`))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		panic(fmt.Errorf("unknown command %s", r.URL.Path))
	})

	handler := http.StripPrefix("/bot"+BotToken, mux)
	handler.ServeHTTP(w, r)
}

const (
	Ivanov     int64 = 256
	Petrov     int64 = 512
	Alexandrov int64 = 1024
	BotChatID        = 100500
)

var (
	users = map[int64]*tgbotapi.User{
		Ivanov: &tgbotapi.User{
			ID:           Ivanov,
			FirstName:    "Ivan",
			LastName:     "Ivanov",
			UserName:     "ivanov",
			LanguageCode: "ru",
			IsBot:        false,
		},
		Petrov: &tgbotapi.User{
			ID:           Petrov,
			FirstName:    "Petr",
			LastName:     "Pertov",
			UserName:     "ppetrov",
			LanguageCode: "ru",
			IsBot:        false,
		},
		Alexandrov: &tgbotapi.User{
			ID:           Alexandrov,
			FirstName:    "Alex",
			LastName:     "Alexandrov",
			UserName:     "aalexandrov",
			LanguageCode: "ru",
			IsBot:        false,
		},
	}

	updID uint64
	msgID uint64
)

func SendMsgToBot(userID int64, text string) error {
	atomic.AddUint64(&updID, 1)
	myUpdID := atomic.LoadUint64(&updID)

	// better have it per user, but lazy now
	atomic.AddUint64(&msgID, 1)
	myMsgID := atomic.LoadUint64(&msgID)

	user, ok := users[userID]
	if !ok {
		return fmt.Errorf("no user for %d", userID)
	}

	upd := &tgbotapi.Update{
		UpdateID: int(myUpdID),
		Message: &tgbotapi.Message{
			MessageID: int(myMsgID),
			From:      user,
			Chat: &tgbotapi.Chat{
				ID:        user.ID,
				FirstName: user.FirstName,
				UserName:  user.UserName,
				Type:      "private",
			},
			Text: text,
			Date: int(time.Now().Unix()),
			Entities: []tgbotapi.MessageEntity{
				{
					Type:   "bot_command",
					Offset: 0,
					Length: len(strings.Split(text, " ")[0]),
				},
			},
		},
	}
	//nolint:errcheck
	reqData, _ := json.Marshal(upd)

	reqBody := bytes.NewBuffer(reqData)
	//nolint:errcheck
	req, _ := http.NewRequest(http.MethodPost, WebhookURL, reqBody)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	//nolint:govet
	defer resp.Body.Close()
	return err
}

type testCase struct {
	user    int64
	command string
	answers map[int64]string
}

func TestTasks(t *testing.T) {
	tds := NewTDS()
	ts := httptest.NewServer(tds)
	tgbotapi.APIEndpoint = ts.URL + "/bot%s/%s"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		err := StartTaskBot(ctx, BotToken, WebhookURL)
		if err != nil {
			//nolint:govet
			t.Fatalf("startTaskBot error: %s", err)
		}
	}()

	// give server time to start
	time.Sleep(100 * time.Millisecond)

	cases := []testCase{
		{
			// команда /tasks - выводит список всех активных задач
			Ivanov,
			"/tasks",
			map[int64]string{
				Ivanov: "Нет задач",
			},
		},
		{
			// команда /new - создаёт новую задачу, всё что после /new - идёт в название задачи
			Ivanov,
			"/new написать бота",
			map[int64]string{
				Ivanov: `Задача "написать бота" создана, id=1`,
			},
		},
		{
			Ivanov,
			"/tasks",
			map[int64]string{
				Ivanov: `1. написать бота by @ivanov
/assign_1`,
			},
		},
		{
			Ivanov,
			"/assign_1",
			map[int64]string{
				Ivanov: `Задача "написать бота" назначена на вас`,
			},
		},
		{
			// /assign_* - назначает задачу на себя
			Alexandrov,
			"/assign_1",
			map[int64]string{
				Alexandrov: `Задача "написать бота" назначена на вас`,
				Ivanov:     `Задача "написать бота" назначена на @aalexandrov`,
			},
		},
		{
			// в случае если задача была назначена на кого-то - он получает уведомление об этом
			// в данном случае она была назначена на Alexandrov, поэтому ему отправляется уведомление
			Petrov,
			"/assign_1",
			map[int64]string{
				Petrov:     `Задача "написать бота" назначена на вас`,
				Alexandrov: `Задача "написать бота" назначена на @ppetrov`,
			},
		},
		{
			// если задача назначена и на мне - показывается "на меня"
			Petrov,
			"/tasks",
			map[int64]string{
				Petrov: `1. написать бота by @ivanov
assignee: я
/unassign_1 /resolve_1`,
			},
		},
		{
			// если задача назначена и не на мне - показывается логин исполнителя
			// при
			Ivanov,
			"/tasks",
			map[int64]string{
				Ivanov: `1. написать бота by @ivanov
assignee: @ppetrov`,
			},
		},

		{
			// /unassign_ - снимает задачу с себя
			// нельзя снять задачу которая не на вас
			Alexandrov,
			"/unassign_1",
			map[int64]string{
				Alexandrov: `Задача не на вас`,
			},
		},

		{
			// /unassign_ - снимает задачу с себя
			// автору отправляется уведомление что задача осталась без исполнителя
			Petrov,
			"/unassign_1",
			map[int64]string{
				Petrov: `Принято`,
				Ivanov: `Задача "написать бота" осталась без исполнителя`,
			},
		},

		{
			// повтор
			// в случае если задача была назначена на кого-то - автор получает уведомление об этом
			Petrov,
			"/assign_1",
			map[int64]string{
				Petrov: `Задача "написать бота" назначена на вас`,
				Ivanov: `Задача "написать бота" назначена на @ppetrov`,
			},
		},
		{
			// /resolve_* завершает задачу, удаляет её из хранилища
			// автору отправляется уведомление об этом
			Petrov,
			"/resolve_1",
			map[int64]string{
				Petrov: `Задача "написать бота" выполнена`,
				Ivanov: `Задача "написать бота" выполнена @ppetrov`,
			},
		},

		{
			Petrov,
			"/tasks",
			map[int64]string{
				Petrov: `Нет задач`,
			},
		},

		{
			// обратите внимание, id=2 - автоинкремент
			Petrov,
			"/new сделать ДЗ по курсу",
			map[int64]string{
				Petrov: `Задача "сделать ДЗ по курсу" создана, id=2`,
			},
		},
		{
			// обратите внимание, id=3 - автоинкремент
			Ivanov,
			"/new прийти на хакатон",
			map[int64]string{
				Ivanov: `Задача "прийти на хакатон" создана, id=3`,
			},
		},
		{
			Petrov,
			"/tasks",
			map[int64]string{
				Petrov: `2. сделать ДЗ по курсу by @ppetrov
/assign_2

3. прийти на хакатон by @ivanov
/assign_3`,
			},
		},
		{
			// повтор
			// в случае если задача была назначена на кого-то - автор получает уведомление об этом
			// если он автор задачи - ему не приходит дополнительного уведомления о том что она назначена на кого-то
			Petrov,
			"/assign_2",
			map[int64]string{
				Petrov: `Задача "сделать ДЗ по курсу" назначена на вас`,
			},
		},
		{
			Petrov,
			"/tasks",
			map[int64]string{
				Petrov: `2. сделать ДЗ по курсу by @ppetrov
assignee: я
/unassign_2 /resolve_2

3. прийти на хакатон by @ivanov
/assign_3`,
			},
		},
		{
			// /my показывает задачи которые назначены на меня
			// при этом тут нет метки assignee
			Petrov,
			"/my",
			map[int64]string{
				Petrov: `2. сделать ДЗ по курсу by @ppetrov
/unassign_2 /resolve_2`,
			},
		},
		{
			// /owner - показывает задачи, которы я создал
			// при этом тут нет метки assignee
			Ivanov,
			"/owner",
			map[int64]string{
				Ivanov: `3. прийти на хакатон by @ivanov
/assign_3`,
			},
		},
	}

	for idx, item := range cases {

		tds.Lock()
		tds.Answers = make(map[int64]string)
		tds.Unlock()

		caseName := fmt.Sprintf("[case%d, %d: %s]", idx, item.user, item.command)
		err := SendMsgToBot(item.user, item.command)
		if err != nil {
			t.Fatalf("%s SendMsgToBot error: %s", caseName, err)
		}
		// give TDS time to process request
		time.Sleep(10 * time.Millisecond)

		tds.Lock()
		result := reflect.DeepEqual(tds.Answers, item.answers)
		if !result {
			t.Fatalf("%s bad results:\n\tWant: %v\n\tHave: %v", caseName, item.answers, tds.Answers)
		}
		tds.Unlock()

	}
}
