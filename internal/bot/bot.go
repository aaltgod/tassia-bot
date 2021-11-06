package bot

import (
	"log"

	postgres "github.com/alyaskastorm/tassia-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	InternalError = "Internal error. Sorry"
)

type Bot struct {
	botApi                *tgbotapi.BotAPI
	usersIDsInInteractive map[int32]string
	storage               postgres.Storage
}

func NewBot(botApi *tgbotapi.BotAPI, usersIDsInInteractive map[int32]string, storage postgres.Storage) *Bot {
	return &Bot{
		botApi:                botApi,
		usersIDsInInteractive: usersIDsInInteractive,
		storage:               storage}
}

func (b *Bot) Start() error {

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botApi.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.CallbackQuery != nil {
			if err := b.handlerCallbackQuery(update.CallbackQuery); err != nil {
				log.Println(err)
				b.handleError(update.Message)
			}

			continue
		}

		if update.Message.IsCommand() {
			if err := b.handleCommand(update.Message); err != nil {
				log.Println(err)
				b.handleError(update.Message)
			}
		}
	}

	return nil
}
