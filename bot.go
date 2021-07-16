package main

import (
	"bytes"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

const (
	InternalError = "Internal error. Sorry"
)

type Handler struct {
	storage Storage
}

func NewHandler(storage Storage) *Handler {
	return &Handler{storage: storage}
}

func RunBot(h *Handler) error {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		return err
	}

	bot.Debug = true

	log.Println("Authorized username: ", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	var userNames = make(map[string]time.Time)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s]: %s\n", update.Message.From.UserName, update.Message.Text)

		var msg tgbotapi.MessageConfig

		switch update.Message.Text {
		case "/df", "/df@GotassiaBot":
			p := exec.Command("df", "-h")

			var out bytes.Buffer
			p.Stdout = &out

			if err := p.Run(); err != nil {
				log.Println(err)

				continue
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, out.String())
		case "/t", "/t@GotassiaBot":
			temp, err := getMoscowTemperature()
			if err != nil {
				log.Println(err)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)

				break
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, temp)
		case "/sleep", "/sleep@GotassiaBot":
			date := update.Message.Time().UTC()
			userName := update.Message.From.UserName
			if _, exists := userNames[userName]; !exists {
				userNames[userName] = date

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Таймер запущен. Сладких снов :)")

				break
			}

			sleepTime := date.Sub(userNames[userName])
			delete(userNames, userName)

			user, err := h.storage.Get(userName)
			if err != nil {
				log.Println("[GET] ", err)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)
				break
			}

			if user.Counter == 0 {
				user.Name = userName
				user.Counter = 1
				user.AverageTimeSleep = sleepTime.Hours()

				err = h.storage.Create(user)
				if err != nil {
					log.Println("[CREATE] ", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)
					break
				}
			} else {
				hoursNumber := user.AverageTimeSleep * float64(user.Counter) + sleepTime.Hours()
				user.Counter++
				user.AverageTimeSleep = hoursNumber / float64(user.Counter)

				if err = h.storage.Update(user); err != nil {
					log.Println("[UPDATE]: ", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)
					break
				}
			}

			msg = tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Ты поспонькал %s", sleepTime.String()),
			)
		case "/sleepstat", "/sleepstat@GotassiaBot":
			if update.Message.Chat.IsChannel() {
				users, err := h.storage.GetAll()
				if err != nil {
					log.Println("{GET-ALL]: ", err)
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)
					break
				}

				result := fmt.Sprintf("%-15s %-10s %-10s\n\n", "[nickname]", "[days]", "[average]")

				for _, user := range users {
					result += fmt.Sprintf("%-15v %8v %10vh\n", user.Name, user.Counter, user.AverageTimeSleep)
				}

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, result)

				break
			}

			userName := update.Message.From.UserName

			user, err := h.storage.Get(userName)
			if err != nil {
				log.Println("{GET]: ", err)
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, InternalError)
				break
			}

			if user.Counter == 0 {
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет статистики")
				break
			}

			result := fmt.Sprintf("%-15s %-10s %-10s\n\n", "[nickname]", "[days]", "[average]")
			result += fmt.Sprintf("%-15v %8v %10vh\n", user.Name, user.Counter, user.AverageTimeSleep)
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, result)
		default:
			continue
		}

		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}

	return nil
}

func getMoscowTemperature() (string, error) {

	resp, err := http.Get("https://yandex.ru/pogoda/?lat=55.85489273&lon=37.47623444")
	if err != nil {
		log.Println(err)

		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Println("response status code ", resp.StatusCode)

		return "", fmt.Errorf("%s", "error response")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Println(err)

		return "", err
	}

	result := doc.Find("div.link__feelings.fact__feelings").Find(
		"div.link__condition.day-anchor").Text()
	result += "\nТекущая температура: " + doc.Find("div.temp.fact__temp.fact__temp_size_s").Find(
		"span.temp__value.temp__value_with-unit").Text() + "°"
	result += "\nОщущается как: " + doc.Find("div.link__feelings.fact__feelings").Find(
		"span.temp__value.temp__value_with-unit").Text() + "°"

	return result, nil
}

