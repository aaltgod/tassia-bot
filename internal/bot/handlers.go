package bot

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	constant "github.com/aaltgod/tassia-bot/internal/constants"
	msgconstructor "github.com/aaltgod/tassia-bot/internal/message-constructor"
	postgres "github.com/aaltgod/tassia-bot/internal/storage"
	sunsetsunrise "github.com/aaltgod/tassia-bot/pkg/sunset-sunrise"
	"github.com/aaltgod/tassia-bot/pkg/temperature"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func (b *Bot) handleCommand(message *tgbotapi.Message) error {
	log.Printf(
		"[%s] with id [%d]: %s\n",
		message.From.UserName,
		message.From.ID,
		message.Text,
	)

	var (
		msg tgbotapi.MessageConfig
	)

	switch message.Command() {
	case "chat":
		var found bool

		for _, id := range b.chatGPT.VerifiedUserIDs {
			if id == message.From.ID {
				found = true

				break
			}
		}

		var result string

		if found {
			userMessage := message.Text

			log.Printf("GOT user [%s] message: %s", message.From.UserName, userMessage)

			<-b.ticker.C

			response, err := b.chatGPT.SendMessage(userMessage)
			if err != nil {
				log.Println(err)

				result = "InternalError"
			} else {
				result = response
			}

			b.ticker.Reset(b.tickerTimeDuration)

			log.Printf("GOT user [%s] result: %s", message.From.UserName, result)
		} else {
			result = "No access"
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, result)
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
	case "i":
		sunIntervals, err := sunsetsunrise.GetSunIntervals()
		if err != nil {
			return err
		}

		msg = tgbotapi.NewMessage(message.Chat.ID, sunIntervals)
	case "sleep":
		userName := message.From.UserName
		chatID := message.Chat.ID

		exists, err := b.dateStorage.DateIsExist(b.ctx, userName, chatID)
		if err != nil {
			return err
		}

		date := message.Time().UTC()

		if !exists {
			user := new(postgres.Date)
			user.Name = userName
			user.StartDate = date.Format(constant.LAYOUT)
			user.ChatID = chatID

			if err = b.dateStorage.CreateStartDate(b.ctx, user); err != nil {
				return err
			}

			msg = tgbotapi.NewMessage(message.Chat.ID, "Таймер запущен. Сладких снов :)")
			break
		}

		user, err := b.dateStorage.GetDate(b.ctx, userName, chatID)
		if err != nil {
			return err
		}

		if user.StopDate != "" {
			msg = tgbotapi.NewMessage(chatID, "Выбери действие у последнего таймера")
			break
		}

		startTime, err := time.Parse(constant.LAYOUT, user.StartDate)
		if err != nil {
			return err
		}

		user.StopDate = date.Format(constant.LAYOUT)
		if err := b.dateStorage.UpdateStopDate(b.ctx, user); err != nil {
			return err
		}

		sleepTime := date.Sub(startTime)

		msg = tgbotapi.NewMessage(
			message.Chat.ID,
			fmt.Sprintf("Ты поспонькал %s", sleepTime.String()),
		)
		buttons := tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Add to stat", constant.ADDTOSTAT),
			tgbotapi.NewInlineKeyboardButtonData("Don't add", constant.DONTADD),
		)
		keyboard := tgbotapi.NewInlineKeyboardMarkup(buttons)
		msg.ReplyMarkup = keyboard
	case "sleepstat":
		if message.Chat.IsGroup() {
			users, err := b.statStorage.GetStats(b.ctx, message.Chat.ID)
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

		user, err := b.statStorage.GetStat(b.ctx, userName, message.Chat.ID)
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
	case constant.ADDTOSTAT:
		userName := message.From.UserName
		messageID := message.Message.MessageID
		chatID := message.Message.Chat.ID
		dateUser, err := b.dateStorage.GetDate(b.ctx, userName, chatID)
		if err != nil {
			return err
		}

		startDate, stopDate := dateUser.StartDate, dateUser.StopDate

		startTime, err := time.Parse(constant.LAYOUT, startDate)
		if err != nil {
			return err
		}

		stopTime, err := time.Parse(constant.LAYOUT, stopDate)
		if err != nil {
			return err
		}

		sleepTime := stopTime.Sub(startTime)

		exists, err := b.statStorage.StatIsExists(b.ctx, userName, chatID)
		if err != nil {
			return err
		}
		if !exists {
			stat := new(postgres.Stat)
			stat.Name = userName
			stat.ChatID = message.Message.Chat.ID
			stat.Counter = 1
			stat.AverageTimeSleep = sleepTime.Hours()

			err = b.statStorage.CreateStat(b.ctx, stat)
			if err != nil {
				return err
			}

		} else {
			stat, err := b.statStorage.GetStat(b.ctx, userName, chatID)
			if err != nil {
				return err
			}

			hoursNumber := stat.AverageTimeSleep*float64(stat.Counter) + sleepTime.Hours()
			stat.Counter++
			stat.AverageTimeSleep = hoursNumber / float64(stat.Counter)

			if err = b.statStorage.UpdateStat(b.ctx, stat); err != nil {
				return err
			}
		}

		if err = b.dateStorage.DeleteDate(b.ctx, userName, dateUser.ChatID); err != nil {
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
	case constant.DONTADD:
		userName := message.From.UserName
		chatID := message.Message.Chat.ID
		dateUser, err := b.dateStorage.GetDate(b.ctx, userName, chatID)
		if err != nil {
			return err
		}

		err = b.dateStorage.DeleteDate(b.ctx, message.From.UserName, dateUser.ChatID)
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
			log.Println("ENTRYPOINT")

			path := message.Data[4:]
			msgConstructor := msgconstructor.NewMessageConstructor(b.dirStorage)
			fileKeyboardRow, err := msgConstructor.CreateFileKeyboarRow(path, 1)
			if err != nil {
				return err
			}

			dirIdx := &msgconstructor.DirIdx{
				First:  0,
				Second: 1,
				Third:  2,
			}

			dirKeyboardRow, err := msgConstructor.CreateDIRKeyboardRow(path, dirIdx)
			if err != nil {
				return err
			}

			text, err := msgConstructor.CreateMessageText(path)
			if err != nil {
				return err
			}

			exists, err := b.dirStorage.DirIsExistsByPath(b.ctx, path)
			if err != nil {
				return err
			}
			if exists {
				_, err := b.dirStorage.GetDirByPath(b.ctx, path)
				if err != nil {
					return err
				}
			}

			dirNavigationKeyboardRow, err := msgConstructor.CreateDIRNavigationButtonsRow(path, path, dirIdx)
			if err != nil {
				return err
			}

			keyboard := tgbotapi.NewInlineKeyboardMarkup(fileKeyboardRow, dirKeyboardRow, dirNavigationKeyboardRow)

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
			log.Println("DIR")
			uuid := message.Data[4:]
			log.Println(message.Data, uuid)
			exists, err := b.dirStorage.DirIsExistsByUUID(b.ctx, uuid)
			if err != nil {
				return err
			}
			if exists {
				dir, err := b.dirStorage.GetDirByUUID(b.ctx, uuid)
				if err != nil {
					return err
				}

				log.Println(dir.Path, dir.ParentPath)
				msgConstructor := msgconstructor.NewMessageConstructor(b.dirStorage)
				fileKeyboardRow, err := msgConstructor.CreateFileKeyboarRow(dir.Path, 1)
				if err != nil {
					return err
				}

				dirIdx := &msgconstructor.DirIdx{
					First:  0,
					Second: 1,
					Third:  2,
				}

				dirKeyboardRow, err := msgConstructor.CreateDIRKeyboardRow(dir.Path, dirIdx)
				if err != nil {
					return err
				}

				text, err := msgConstructor.CreateMessageText(dir.Path)
				if err != nil {
					return err
				}

				dirNavigationKeyboardRow, err := msgConstructor.CreateDIRNavigationButtonsRow(dir.Path, dir.ParentPath, dirIdx)
				if err != nil {
					return err
				}

				log.Println(fileKeyboardRow, dirKeyboardRow, dirNavigationKeyboardRow)

				keyboard := tgbotapi.NewInlineKeyboardMarkup(fileKeyboardRow, dirKeyboardRow, dirNavigationKeyboardRow)

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
			}

		case constant.FLIPDIRS:
			log.Println("FLIPDIRS")
			uuid := message.Data[4 : len(message.Data)-6]
			log.Println(message.Data, uuid)

			exists, err := b.dirStorage.DirIsExistsByUUID(b.ctx, uuid)
			if err != nil {
				return err
			}
			if exists {
				idxs := strings.Split(message.Data[len(message.Data)-5:], ":")
				log.Println(idxs)

				dir, err := b.dirStorage.GetDirByUUID(b.ctx, uuid)
				if err != nil {
					return err
				}

				msgConstructor := msgconstructor.NewMessageConstructor(b.dirStorage)
				fileKeyboardRow, err := msgConstructor.CreateFileKeyboarRow(dir.Path, 1)
				if err != nil {
					return err
				}

				firstIdx, err := strconv.Atoi(idxs[0])
				if err != nil {
					return err
				}
				secondIdx, err := strconv.Atoi(idxs[1])
				if err != nil {
					return err
				}
				thirdIdx, err := strconv.Atoi(idxs[2])
				if err != nil {
					return err
				}

				dirIdx := &msgconstructor.DirIdx{
					First:  firstIdx,
					Second: secondIdx,
					Third:  thirdIdx,
				}

				dirKeyboardRow, err := msgConstructor.CreateDIRKeyboardRow(dir.Path, dirIdx)
				if err != nil {
					return err
				}

				dirNavigationKeyboardRow, err := msgConstructor.CreateDIRNavigationButtonsRow(dir.Path, dir.ParentPath, dirIdx)
				if err != nil {
					return err
				}

				keyboard := tgbotapi.NewInlineKeyboardMarkup(fileKeyboardRow, dirKeyboardRow, dirNavigationKeyboardRow)

				editMessageReplyMarkup := tgbotapi.NewEditMessageReplyMarkup(
					message.Message.Chat.ID,
					message.Message.MessageID,
					keyboard,
				)

				b.botApi.Send(editMessageReplyMarkup)
			}

		case constant.FILE:

		}
	}

	return nil
}
