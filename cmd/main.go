package main

import (
	"log"
	"os"
	"time"

	b "github.com/alyaskastorm/tassia-bot/internal/bot"
	postgres "github.com/alyaskastorm/tassia-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
)

func main() {
	log.Println("Start")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second * 3)

	db, err := postgres.CreateConnection()
	if err != nil {
		log.Fatalln(err)
	}

	if err = postgres.PrepareStorage(db); err != nil {
		log.Fatalln(err)
	}
	db.Close()

	storage := postgres.NewUserStorage()

	botApi, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	botApi.Debug = true

	log.Println("Authorized username: ", botApi.Self.UserName)

	usersIdsInInteractive := make(map[int32]string)
	bot := b.NewBot(botApi, usersIdsInInteractive, storage)

	if err = bot.Start(); err != nil {
		log.Fatalln(err)
	}
}
