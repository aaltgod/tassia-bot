package main

import (
	"github.com/joho/godotenv"
	"log"
)

func main() {
	log.Println("Start")

	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	db, err := CreateConnection()
	if err != nil {
		log.Fatalln(err)
	}

	err = PrepareStorage(db)
	if err != nil {
		log.Fatalln(err)
	}
	db.Close()

	storage := NewUserStorage()
	handler := NewHandler(storage)

	err = RunBot(handler)
	if err != nil {
		log.Fatalln(err)
	}
}
