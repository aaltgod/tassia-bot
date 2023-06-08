package bot

import (
	"context"
	"log"
	"sync"
	"time"

	chatGPT "github.com/aaltgod/tassia-bot/internal/chat-gpt"
	postgres "github.com/aaltgod/tassia-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	InternalError = "Internal error. Sorry"
)

type Bot struct {
	ctx context.Context

	ticker             *time.Ticker
	tickerTimeDuration time.Duration

	chatGPT     *chatGPT.Client
	botApi      *tgbotapi.BotAPI
	statStorage postgres.StatStorage
	dateStorage postgres.DateStorage
	dirStorage  postgres.DirStorage
}

func NewBot(
	ctx context.Context,
	tickerTimeDuration time.Duration,
	chatGPT *chatGPT.Client,
	botApi *tgbotapi.BotAPI,
	statStorage postgres.StatStorage,
	dateStorage postgres.DateStorage,
	dirStorage postgres.DirStorage,
) *Bot {
	return &Bot{
		ctx:                ctx,
		ticker:             time.NewTicker(tickerTimeDuration),
		tickerTimeDuration: tickerTimeDuration,
		chatGPT:            chatGPT,
		botApi:             botApi,
		statStorage:        statStorage,
		dateStorage:        dateStorage,
		dirStorage:         dirStorage,
	}
}

func (b *Bot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := b.botApi.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	var (
		wg        = &sync.WaitGroup{}
		semaphore = make(chan struct{}, 6)
	)

	for update := range updates {
		if update.Message == nil && update.CallbackQuery == nil {
			continue
		}

		if update.CallbackQuery != nil {
			semaphore <- struct{}{}

			wg.Add(1)

			go func(update tgbotapi.Update) {
				if err := b.handlerCallbackQuery(update.CallbackQuery); err != nil {
					log.Println(err)
					b.handleError(update.Message)
				}

				wg.Done()

				<-semaphore
			}(update)

			continue
		}

		if update.Message.IsCommand() {
			semaphore <- struct{}{}

			wg.Add(1)

			go func(update tgbotapi.Update) {
				if err := b.handleCommand(update.Message); err != nil {
					log.Println(err)
					b.handleError(update.Message)
				}

				wg.Done()

				<-semaphore
			}(update)
		}
	}

	wg.Wait()

	return nil
}
