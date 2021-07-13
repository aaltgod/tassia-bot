package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Start")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Println("Authorized username: ", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	var usersIDS = make(map[int]time.Time)

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
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Internal error. Sorry")

				break
			}

			msg = tgbotapi.NewMessage(update.Message.Chat.ID, temp)
		case "/sleep", "/sleep@GotassiaBot":
			date := update.Message.Time().UTC()

			userID := update.Message.From.ID
			if _, exists := usersIDS[userID]; !exists  {
				usersIDS[userID] = date

				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Таймер запущен. Сладких снов :)")

				break
			}

			sleepTime := date.Sub(usersIDS[userID])
			delete(usersIDS, userID)

			msg = tgbotapi.NewMessage(
				update.Message.Chat.ID,
				fmt.Sprintf("Ты поспонькал %s", sleepTime.String()),
				)
		default:
			continue
		}

		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
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

	result := doc.Find("div.link__feelings.fact__feelings").Find("div.link__condition.day-anchor").Text()
	result += "\nТекущая температура: " + doc.Find("div.temp.fact__temp.fact__temp_size_s").Find("span.temp__value.temp__value_with-unit").Text() + "°"
	result += "\nОщущается как: " + doc.Find("div.link__feelings.fact__feelings").Find("span.temp__value.temp__value_with-unit").Text() + "°"

	return result, nil
}
