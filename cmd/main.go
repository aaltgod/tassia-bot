package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"

	b "github.com/alyaskastorm/tassia-bot/internal/bot"
	postgres "github.com/alyaskastorm/tassia-bot/internal/storage"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/extra/bundebug"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
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

	// time.Sleep(time.Second * 3)

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

	if err = postgres.PrepareStorage(ctx, db); err != nil {
		log.Fatalln(err)
	}

	storage := postgres.NewStorage(db)

	botApi, err := tgbotapi.NewBotAPI(os.Getenv("API_TOKEN"))
	if err != nil {
		log.Fatalln(err)
	}

	botApi.Debug = true

	log.Println("Authorized username: ", botApi.Self.UserName)

	bot := b.NewBot(ctx, botApi, storage, storage)

	if err = bot.Start(); err != nil {
		log.Fatalln(err)
	}
}
