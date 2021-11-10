package bot

import (
	"context"
	"log"

	postgres "github.com/alyaskastorm/tassia-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	InternalError = "Internal error. Sorry"
)

type Bot struct {
	ctx         context.Context
	botApi      *tgbotapi.BotAPI
	statStorage postgres.StatStorage
	dateStorage postgres.DateStorage
}

func NewBot(ctx context.Context, botApi *tgbotapi.BotAPI, statStorage postgres.StatStorage, dateStorage postgres.DateStorage) *Bot {
	return &Bot{
		ctx:         ctx,
		botApi:      botApi,
		statStorage: statStorage,
		dateStorage: dateStorage,
	}
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
