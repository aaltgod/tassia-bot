package main

import (
	"bytes"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"log"
	"os/exec"
	"time"
)

const (
	addToStat = "Add to stat"
	dontAdd = "Don't add"
	layout = "Jan 2, 2006 at 3:04:05pm"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	log.Printf("[%s]: %s\n", message.From.UserName, message.Text)

	var (
		msg tgbotapi.MessageConfig
	)

	switch message.Command(){
	case "df":
		p := exec.Command("df", "-h")

		var out bytes.Buffer
		p.Stdout = &out

		if err := p.Run(); err != nil {
			log.Println(err)
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, out.String())
	case "t":
		temp, err := getMoscowTemperature()
		if err != nil {
			return err
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, temp)
	case "sleep":
		date := message.Time().UTC()
		userName := message.From.UserName
		user, err := b.storage.GetDate(userName)
		if err != nil {
			return err
		}

		if user.Start == "" {
			user.Start = date.Format(layout)

			if err = b.storage.CreateStartDate(user); err != nil {
				return err
			}

			msg = tgbotapi.NewMessage(message.Chat.ID, "Таймер запущен. Сладких снов :)")
			break
		}

		startTime, err := time.Parse(layout, user.Start)
		if err != nil {
			return err
		}

		user.Stop = date.Format(layout)
		if err := b.storage.UpdateStopDate(user); err != nil {
			return err
		}

		sleepTime := date.Sub(startTime)

		msg = tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ты поспонькал %s", sleepTime.String()),
		)

		button := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add to stat", addToStat),
			tgbotapi.NewInlineKeyboardButtonData("Don't add", dontAdd),
			)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(button)
		msg.ReplyMarkup = keyboard
	case "sleepstat":
		if message.Chat.IsChannel() {
			users, err := b.storage.GetStats()
			if err != nil {
				return err
			}

			result := fmt.Sprintf("%-15s %-10s %-10s\n\n", "[nickname]", "[times]", "[average]")

			for _, user := range users {
				result += fmt.Sprintf("%-15v %8v %10vh\n", user.Name, user.Counter, user.AverageTimeSleep)
			}

			msg = tgbotapi.NewMessage(message.Chat.ID, result)
			break
		}

		userName := message.From.UserName

		user, err := b.storage.GetStat(userName)
		if err != nil {
			return err
		}

		if user.Counter == 0 {
			msg = tgbotapi.NewMessage(message.Chat.ID, "У вас нет статистики")
			break
		}

		result := fmt.Sprintf("%-15s %-10s %-10s\n\n", "[nickname]", "[times]", "[average]")
		result += fmt.Sprintf("%-15v %8v %10vh\n", user.Name, user.Counter, user.AverageTimeSleep)
		msg = tgbotapi.NewMessage(message.Chat.ID, result)
	default:
		return nil
	}

	msg.ReplyToMessageID = message.MessageID
	b.botApi.Send(msg)
	return nil
}

func (b *Bot) handleError(message *tgbotapi.Message) {
	msg := tgbotapi.NewMessage(message.Chat.ID, InternalError)
	b.botApi.Send(msg)
}

func (b *Bot) handlerCallbackQuery(message *tgbotapi.CallbackQuery) error {

	switch message.Data {
	case addToStat:
		userName := message.From.UserName
		dateUser, err := b.storage.GetDate(userName)
		if err != nil {
			return err
		}

		startDate, stopDate := dateUser.Start, dateUser.Stop

		startTime, err := time.Parse(layout, startDate)
		if err != nil {
			return err
		}

		stopTime, err := time.Parse(layout, stopDate)
		if err != nil {
			return err
		}

		sleepTime := stopTime.Sub(startTime)

		statUser, err := b.storage.GetStat(userName)
		if err != nil {
			return err
		}

		if statUser.Counter == 0 {
			statUser.Name = userName
			statUser.Counter = 1
			statUser.AverageTimeSleep = sleepTime.Hours()

			err = b.storage.CreateStat(statUser)
			if err != nil {
				return err
			}
		} else {
			hoursNumber := statUser.AverageTimeSleep * float64(statUser.Counter) + sleepTime.Hours()
			statUser.Counter++
			statUser.AverageTimeSleep = hoursNumber / float64(statUser.Counter)

			if err = b.storage.UpdateStat(statUser); err != nil {
				return err
			}
		}

		msg := tgbotapi.CallbackConfig{
			CallbackQueryID: message.ID,
			Text: "Добавлено",
		}

		if err = b.storage.DeleteDate(userName); err != nil {
			return err
		}

		b.botApi.AnswerCallbackQuery(msg)
		b.botApi.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:message.Message.Chat.ID,
			MessageID: message.Message.MessageID,
		})
		b.botApi.Send(tgbotapi.NewMessage(message.Message.Chat.ID, message.Message.Text))
	case dontAdd:
		if err := b.storage.DeleteDate(message.From.UserName); err != nil {
			return err
		}

		b.botApi.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:message.Message.Chat.ID,
			MessageID: message.Message.MessageID,
		})
		b.botApi.Send(tgbotapi.NewMessage(message.Message.Chat.ID, message.Message.Text))
	}

	return nil
}