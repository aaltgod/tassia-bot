package botkeyboard

import (
	"fmt"
	"log"
	"math"
	"os"

	constant "github.com/alyaskastorm/tassia-bot/pkg/constants"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

func CreateDIRKeyboardRow(path string, dirName string) ([]tgbotapi.InlineKeyboardButton, error) {

	dir, err := os.ReadDir(path)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	keyboardRow := tgbotapi.NewInlineKeyboardRow()
	// keyboardRow = append(
	// 	keyboardRow,
	// 	tgbotapi.NewInlineKeyboardButtonData(
	// 		"◀️",
	// 		fmt.Sprintf("%s archive", DIR),
	// 	))

	for _, v := range dir {
		if v.IsDir() {
			if len(keyboardRow) == 3 {
				break
			}
			data := fmt.Sprintf(
				"%s %s/%s",
				constant.DIR, path, v.Name(),
			)
			log.Println(len(data))
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					v.Name(),
					data,
				),
			)
		}
	}

	// keyboardRow = append(
	// 	keyboardRow,
	// 	tgbotapi.NewInlineKeyboardButtonData(
	// 		"▶️",
	// 		fmt.Sprintf("%s archive", DIR),
	// 	),
	// )

	return keyboardRow, nil

}

func CreateFileKeyboarRow(path string, pageNumber int) ([]tgbotapi.InlineKeyboardButton, error) {

	dir, err := os.ReadDir(path)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	var (
		files       []int
		keyboardRow = tgbotapi.NewInlineKeyboardRow()
	)

	for i, v := range dir {
		if !v.IsDir() {
			files = append(files, i)
		}
	}

	pagesAmount := int(math.Ceil(float64(len(files)) / 5))
	pagesNumbers := make([]int, 0, pagesAmount)

	pagesNumbers = append(pagesNumbers, 1)

	if pagesAmount == 0 {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	if pageNumber <= 4 {
		for i := 1; i < pagesAmount-1; i++ {
			pagesNumbers = append(pagesNumbers, i)
		}
		pagesNumbers = append(pagesNumbers, pagesAmount)
	} else if pageNumber > 4 && pageNumber < pagesAmount-2 {
		pagesNumbers = append(pagesNumbers, pageNumber-1)
		pagesNumbers = append(pagesNumbers, pageNumber)
		pagesNumbers = append(pagesNumbers, pageNumber+1)
		pagesNumbers = append(pagesNumbers, pagesAmount)
	} else {
		for i := 3; i >= 0; i-- {
			pagesNumbers = append(pagesNumbers, pagesAmount-i)
		}
	}

	for _, number := range pagesNumbers {
		var text string

		if number == pageNumber {
			text = fmt.Sprintf("[%d]", number)
		} else {
			text = fmt.Sprintf("%d", number)
		}

		keyboardRow = append(
			keyboardRow,
			tgbotapi.NewInlineKeyboardButtonData(
				text,
				fmt.Sprintf(
					"%s %s %d",
					constant.FILE, path, number,
				),
			),
		)
	}

	return keyboardRow, nil
}
