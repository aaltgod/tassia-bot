package main

import (
	"bytes"
	"log"
	"os"
	"os/exec"

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

	log.Printf("Authorized useraname: %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	for update := range updates {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s]: %s", update.Message.From.UserName, update.Message.Text)

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
		default:
			msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Hi there")
		}

		msg.ReplyToMessageID = update.Message.MessageID

		bot.Send(msg)
	}
}
