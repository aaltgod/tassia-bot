package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"

	b "github.com/aaltgod/tassia-bot/internal/bot"
	chatGPT "github.com/aaltgod/tassia-bot/internal/chat-gpt"
	postgres "github.com/aaltgod/tassia-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"
)

func main() {
	log.Println("Start")

	var (
		ctx          = context.Background()
		maxOpenConns = 4 * runtime.GOMAXPROCS(0)
	)

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	dsn := fmt.Sprintf(
		"postgresql://%s:%s@localhost:%s/storage?sslmode=disable",
		os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_EXTERNAL_PORT"),
	)
	pgdb := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(dsn)))
	db := bun.NewDB(pgdb, pgdialect.New())
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	db.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	db.SetMaxOpenConns(maxOpenConns)
	db.SetMaxIdleConns(maxOpenConns)

	storage := postgres.NewStorage(db)

	if err = storage.PrepareTables(ctx); err != nil {
		log.Fatalln(err)
	}
	if err := storage.PrepareArchiveTable(ctx); err != nil {
		log.Fatal(err)
	}

	botApi, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	botApi.Debug = false

	log.Println("Authorized username: ", botApi.Self.UserName)

	verifiedUserIDs := func() []int {
		var result []int
        for _, v := range strings.Split(os.Getenv("VERIFIED_USER_IDS_FOR_CHAT_GPT"), ",") {
			id, err := strconv.Atoi(v)
			if err != nil {
                log.Fatal("couldn't pasrse VERIFIED_USER_IDS_FOR_CHAT_GPT ", err.Error())
			}

			result = append(result, id)
		}

		return result
	}()

    log.Println("VERIFIED_USER_IDS_FOR_CHAT_GPT ", verifiedUserIDs)

	bot := b.NewBot(
		ctx,
		chatGPT.New(os.Getenv("CHAT_GPT_TOKEN"), verifiedUserIDs),
		botApi,
		storage,
		storage,
		storage,
	)

	if err = bot.Start(); err != nil {
		log.Fatalln(err)
	}
}
