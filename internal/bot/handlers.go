package bot

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	botkeyboard "github.com/alyaskastorm/tassia-bot/pkg/bot-keyboard"
	constant "github.com/alyaskastorm/tassia-bot/pkg/constants"
	temperature "github.com/alyaskastorm/tassia-bot/pkg/temperature"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	log.Printf("[%s]: %s\n", message.From.UserName, message.Text)

	var (
		msg tgbotapi.MessageConfig
	)

	switch message.Command() {
	case "df":
		p := exec.Command("df", "-h")

		var out bytes.Buffer
		p.Stdout = &out

		if err := p.Run(); err != nil {
			log.Println(err)
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, out.String())
	case "t":
		temp, err := temperature.GetMoscowTemperature()
		if err != nil {
			return err
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, temp)
	case "sleep":
		userName := message.From.UserName
		chatID := message.Chat.ID
		user, err := b.storage.GetDate(userName, chatID)
		if err != nil {
			return err
		}

		date := message.Time().UTC()

		if user.Start == "" {
			user.Start = date.Format(constant.Layout)
			user.ChatID = chatID

			if err = b.storage.CreateStartDate(user); err != nil {
				return err
			}

			msg = tgbotapi.NewMessage(message.Chat.ID, "Таймер запущен. Сладких снов :)")
			break
		}

		if user.Stop != "" {
			msg = tgbotapi.NewMessage(chatID, "Выбери действие у последнего таймера")
			break
		}

		startTime, err := time.Parse(constant.Layout, user.Start)
		if err != nil {
			return err
		}

		user.Stop = date.Format(constant.Layout)
		if err := b.storage.UpdateStopDate(user); err != nil {
			return err
		}

		sleepTime := date.Sub(startTime)

		msg = tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ты поспонькал %s", sleepTime.String()),
		)
		buttons := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add to stat", constant.AddToStat),
			tgbotapi.NewInlineKeyboardButtonData("Don't add", constant.DontAdd),
		)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
		msg.ReplyMarkup = keyboard
	case "sleepstat":
		if message.Chat.IsGroup() {
			users, err := b.storage.GetStats(message.Chat.ID)
			if err != nil {
				return err
			}

			if len(users) == 0 {
				msg = tgbotapi.NewMessage(message.Chat.ID, "У этой группы нет статистики")
				break
			}

			result := fmt.Sprintf("%-15s %-10s %-10s\n\n", "[nickname]", "[times]", "[average]")

			for _, user := range users {
				result += fmt.Sprintf("%-15v %8v %10vh\n", user.Name, user.Counter, user.AverageTimeSleep)
			}

			msg = tgbotapi.NewMessage(message.Chat.ID, result)
			break
		}

		userName := message.From.UserName

		user, err := b.storage.GetStat(userName, message.Chat.ID)
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
	case "archive":
		buttons := tgbotapi.NewInlineKeyboardRow()

		dir, err := os.ReadDir("archive")
		if err != nil {
			return err
		}

		for _, v := range dir {
			if v.IsDir() {
				buttons = append(
					buttons,
					tgbotapi.NewInlineKeyboardButtonData(
						v.Name(),
						fmt.Sprintf(
							"%s archive/%s",
							constant.ENTRYPOINT, v.Name(),
						),
					),
				)
			}
		}
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
		msg = tgbotapi.NewMessage(
			message.Chat.ID,
			"*Выбери архив*",
		)
		msg.ParseMode = tgbotapi.ModeMarkdown
		msg.ReplyMarkup = keyboard
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
	case constant.AddToStat:
		userName := message.From.UserName
		messageID := message.Message.MessageID
		chatID := message.Message.Chat.ID
		dateUser, err := b.storage.GetDate(userName, chatID)
		if err != nil {
			return err
		}

		startDate, stopDate := dateUser.Start, dateUser.Stop

		startTime, err := time.Parse(constant.Layout, startDate)
		if err != nil {
			return err
		}

		stopTime, err := time.Parse(constant.Layout, stopDate)
		if err != nil {
			return err
		}

		sleepTime := stopTime.Sub(startTime)

		statUser, err := b.storage.GetStat(userName, chatID)
		if err != nil {
			return err
		}

		if statUser.Counter == 0 {
			statUser.Name = userName
			statUser.ChatID = message.Message.Chat.ID
			statUser.Counter = 1
			statUser.AverageTimeSleep = sleepTime.Hours()

			err = b.storage.CreateStat(statUser)
			if err != nil {
				return err
			}
		} else {
			hoursNumber := statUser.AverageTimeSleep*float64(statUser.Counter) + sleepTime.Hours()
			statUser.Counter++
			statUser.AverageTimeSleep = hoursNumber / float64(statUser.Counter)

			if err = b.storage.UpdateStat(statUser); err != nil {
				return err
			}
		}

		if err = b.storage.DeleteDate(userName, dateUser.ChatID); err != nil {
			return err
		}

		b.botApi.AnswerCallbackQuery(tgbotapi.CallbackConfig{
			CallbackQueryID: message.ID,
			Text:            "Добавлено",
		})
		b.botApi.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    chatID,
			MessageID: messageID,
		})

		msg := tgbotapi.NewMessage(chatID, message.Message.Text)
		msg.ReplyToMessageID = message.Message.ReplyToMessage.MessageID
		b.botApi.Send(msg)
	case constant.DontAdd:
		userName := message.From.UserName
		chatID := message.Message.Chat.ID
		dateUser, err := b.storage.GetDate(userName, chatID)
		if err != nil {
			return err
		}

		err = b.storage.DeleteDate(message.From.UserName, dateUser.ChatID)
		if err != nil {
			return err
		}

		b.botApi.DeleteMessage(tgbotapi.DeleteMessageConfig{
			ChatID:    chatID,
			MessageID: message.Message.MessageID,
		})

		msg := tgbotapi.NewMessage(chatID, message.Message.Text)
		msg.ReplyToMessageID = message.Message.ReplyToMessage.MessageID
		b.botApi.Send(msg)
	default:
		switch message.Data[0:3] {
		case constant.ENTRYPOINT:
			path := message.Data[4:]
			log.Println(path)

			fileKeyboardRow, err := botkeyboard.CreateFileKeyboarRow(path, 1)
			if err != nil {
				return err
			}

			dirKeyboardRow, err := botkeyboard.CreateDIRKeyboardRow(path, "")
			if err != nil {
				return err
			}

			text := fmt.Sprintf("Количество файлов в данной папке: %d\n\n", len(fileKeyboardRow))

			keyboardRow := tgbotapi.NewInlineKeyboardRow()
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️",
					fmt.Sprintf("%s archive", constant.DIR),
				))
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					"▶️",
					fmt.Sprintf("%s archive", constant.DIR),
				),
			)

			keyboard := tgbotapi.NewInlineKeyboardMarkup(fileKeyboardRow, dirKeyboardRow, keyboardRow)

			editMessageReplyMarkup := tgbotapi.NewEditMessageReplyMarkup(
				message.Message.Chat.ID,
				message.Message.MessageID,
				keyboard,
			)
			editMessageText := tgbotapi.NewEditMessageText(
				message.Message.Chat.ID,
				message.Message.MessageID,
				text,
			)

			b.botApi.Send(editMessageText)
			b.botApi.Send(editMessageReplyMarkup)
		case constant.DIR:
			path := message.Data[4:]

			dirKeyboardRow, err := botkeyboard.CreateDIRKeyboardRow(path, "")
			if err != nil {
				return err
			}

			keyboardRow := tgbotapi.NewInlineKeyboardRow()
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️",
					fmt.Sprintf("%s archive", constant.DIR),
				))
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					"▶️",
					fmt.Sprintf("%s archive", constant.DIR),
				),
			)

			keyboard := tgbotapi.NewInlineKeyboardMarkup(dirKeyboardRow, keyboardRow)

			editMessageReplyMarkup := tgbotapi.NewEditMessageReplyMarkup(
				message.Message.Chat.ID,
				message.Message.MessageID,
				keyboard,
			)

			b.botApi.Send(editMessageReplyMarkup)
		case constant.FILE:

		}
	}

	return nil
}