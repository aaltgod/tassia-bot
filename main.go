package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"log"
	"os"
	"time"
)

func main() {
	log.Println("Start")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 3)

	db, err := CreateConnection()
	if err != nil {
		log.Fatalln(err)
	}

	if err = PrepareStorage(db); err != nil {
		log.Fatalln(err)
	}
	db.Close()

	storage := NewUserStorage()

	botApi, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	botApi.Debug = true

	log.Println("Authorized username: ", botApi.Self.UserName)

	bot := NewBot(botApi, storage)

	if err = bot.Start(); err != nil {
		log.Fatalln(err)
	}
}
