package messageconstructor

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"math"
	"os"
	"strings"

	constant "github.com/alyaskastorm/tassia-bot/internal/constants"

	postgres "github.com/alyaskastorm/tassia-bot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Message interface {
	CreateDIRKeyboardRow(path string, dirIdx *DirIdx) ([]tgbotapi.InlineKeyboardButton, error)
	CreateFileKeyboarRow(path string, pageNumber int) ([]tgbotapi.InlineKeyboardButton, error)
	CreateMessageText(path string) (string, error)
	CreateDIRNavigationButtonsRow(path, parentDirPath string, dirIdx *DirIdx) ([]tgbotapi.InlineKeyboardButton, error)
}

type DirIdx struct {
	First, Second, Third int
}

type MessageConstructor struct {
	DirStorage postgres.DirStorage
}

func NewMessageConstructor(dirStroage postgres.DirStorage) *MessageConstructor {
	return &MessageConstructor{
		DirStorage: dirStroage,
	}
}

func (m *MessageConstructor) CreateDIRKeyboardRow(path string, dirIdx *DirIdx) ([]tgbotapi.InlineKeyboardButton, error) {

	keyboardRow := tgbotapi.NewInlineKeyboardRow()

	dirs, err := GetDirs(path)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	if len(dirs) == 0 {
		return []tgbotapi.InlineKeyboardButton{}, nil
	}

	keyboardRow = append(
		keyboardRow,
		tgbotapi.NewInlineKeyboardButtonData(
			dirs[dirIdx.First].Name(),
			fmt.Sprintf(
				"%s/%s",
				path,
				dirs[dirIdx.First].Name(),
			),
		),
	)

	if dirIdx.Second != 0 && dirIdx.Second < len(dirs) {
		keyboardRow = append(
			keyboardRow,
			tgbotapi.NewInlineKeyboardButtonData(
				dirs[dirIdx.Second].Name(),
				fmt.Sprintf(
					"%s/%s",
					path,
					dirs[dirIdx.Second].Name(),
				),
			),
		)
	}

	if dirIdx.Third != 0 && dirIdx.Third < len(dirs) {
		keyboardRow = append(
			keyboardRow,
			tgbotapi.NewInlineKeyboardButtonData(
				dirs[dirIdx.Third].Name(),
				fmt.Sprintf(
					"%s/%s",
					path,
					dirs[dirIdx.Third].Name(),
				),
			),
		)
	}

	ctx := context.Background()

	for i, button := range keyboardRow {
		exists, err := m.DirStorage.DirIsExistsByPath(ctx, *button.CallbackData)
		if err != nil {
			return []tgbotapi.InlineKeyboardButton{}, err
		}
		if !exists {
			if i < len(keyboardRow)-1 {
				copy(keyboardRow[:i], keyboardRow[i+1:])
			} else {
				keyboardRow = keyboardRow[:len(keyboardRow)-1]
			}

		} else {
			dir, err := m.DirStorage.GetDirByPath(ctx, *button.CallbackData)
			if err != nil {
				return []tgbotapi.InlineKeyboardButton{}, err
			}

			*button.CallbackData = fmt.Sprintf("%s %s", constant.DIR, dir.UUID)
			keyboardRow[i] = button
		}
	}

	return keyboardRow, nil
}

func (m *MessageConstructor) CreateFileKeyboarRow(path string, pageNumber int) ([]tgbotapi.InlineKeyboardButton, error) {

	dir, err := os.ReadDir(path)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	var (
		filesIdxs   []int
		keyboardRow = tgbotapi.NewInlineKeyboardRow()
	)

	for i, v := range dir {
		if !v.IsDir() {
			filesIdxs = append(filesIdxs, i)
		}
	}

	pagesAmount := int(math.Ceil(float64(len(filesIdxs)) / 5))
	pagesNumbers := make([]int, 0, pagesAmount)
	pagesNumbers = append(pagesNumbers, 1)

	if pagesAmount <= 1 {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	if pageNumber <= 4 {
		for i := 2; i < pagesAmount-1; i++ {
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

func (m *MessageConstructor) CreateMessageText(path string) (string, error) {

	files, err := GetFiles(path)
	if err != nil {
		return "", err
	}

	dirNames := strings.Split(path, "/")
	dirName := dirNames[len(dirNames)-1]

	text := fmt.Sprintf(
		"Количество файлов в [%s]: %d\n\n",
		dirName,
		len(files),
	)

	for _, file := range files {
		text += file.Name() + "\n\n"
	}

	return text, nil
}

func (m *MessageConstructor) CreateDIRNavigationButtonsRow(path, parentDirPath string, dirIdx *DirIdx) ([]tgbotapi.InlineKeyboardButton, error) {

	log.Println(dirIdx)
	keyboardRow := tgbotapi.NewInlineKeyboardRow()
	ctx := context.Background()

	if dirIdx.First > 2 {
		exists, err := m.DirStorage.DirIsExistsByPath(ctx, path)
		if err != nil {
			return []tgbotapi.InlineKeyboardButton{}, err
		}
		if exists {
			dir, err := m.DirStorage.GetDirByPath(ctx, path)
			if err != nil {
				return []tgbotapi.InlineKeyboardButton{}, err
			}
			keyboardRow = append(
				keyboardRow,
				tgbotapi.NewInlineKeyboardButtonData(
					"◀️",
					fmt.Sprintf(
						"%s %s %d:%d:%d",
						constant.FLIPDIRS, dir.UUID,
						dirIdx.First-3,
						dirIdx.First-2,
						dirIdx.First-1,
					),
				),
			)
		}
	}

	exists, err := m.DirStorage.DirIsExistsByPath(ctx, parentDirPath)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}
	if exists {
		dir, err := m.DirStorage.GetDirByPath(ctx, parentDirPath)
		if err != nil {
			return []tgbotapi.InlineKeyboardButton{}, err
		}

		keyboardRow = append(
			keyboardRow,
			tgbotapi.NewInlineKeyboardButtonData(
				"🔽",
				fmt.Sprintf(
					"%s %s",
					constant.DIR,
					dir.UUID),
			),
		)
	}

	dirs, err := GetDirs(path)
	if err != nil {
		return []tgbotapi.InlineKeyboardButton{}, err
	}

	dirsAmount := len(dirs)

	if dirIdx.Third < dirsAmount-1 && dirIdx.Third != 0 {
		nextDirsAmount := dirsAmount - 1 - dirIdx.Third
		log.Println("AMOUNT", nextDirsAmount, dirsAmount, dirIdx.Third)
		switch nextDirsAmount {
		case 1:
			exists, err := m.DirStorage.DirIsExistsByPath(ctx, path)
			if err != nil {
				return []tgbotapi.InlineKeyboardButton{}, err
			}
			if exists {
				dir, err := m.DirStorage.GetDirByPath(ctx, path)
				if err != nil {
					return []tgbotapi.InlineKeyboardButton{}, err
				}
				keyboardRow = append(
					keyboardRow,
					tgbotapi.NewInlineKeyboardButtonData(
						"▶️",
						fmt.Sprintf(
							"%s %s %d:%d:%d",
							constant.FLIPDIRS, dir.UUID,
							dirIdx.Third+1,
							0,
							0,
						),
					),
				)
			}
		case 2:
			exists, err := m.DirStorage.DirIsExistsByPath(ctx, path)
			if err != nil {
				return []tgbotapi.InlineKeyboardButton{}, err
			}
			if exists {
				dir, err := m.DirStorage.GetDirByPath(ctx, path)
				if err != nil {
					return []tgbotapi.InlineKeyboardButton{}, err
				}
				keyboardRow = append(
					keyboardRow,
					tgbotapi.NewInlineKeyboardButtonData(
						"▶️",
						fmt.Sprintf(
							"%s %s %d:%d:%d",
							constant.FLIPDIRS, dir.UUID,
							dirIdx.Third+1,
							dirIdx.Third+2,
							0,
						),
					),
				)
			}
		default:
			exists, err := m.DirStorage.DirIsExistsByPath(ctx, path)
			if err != nil {
				return []tgbotapi.InlineKeyboardButton{}, err
			}
			if exists {
				dir, err := m.DirStorage.GetDirByPath(ctx, path)
				if err != nil {
					return []tgbotapi.InlineKeyboardButton{}, err
				}
				keyboardRow = append(
					keyboardRow,
					tgbotapi.NewInlineKeyboardButtonData(
						"▶️",
						fmt.Sprintf(
							"%s %s %d:%d:%d",
							constant.FLIPDIRS, dir.UUID,
							dirIdx.Third+1,
							dirIdx.Third+2,
							dirIdx.Third+3,
						),
					),
				)
			}
		}
	}

	return keyboardRow, nil
}

func GetFiles(path string) ([]fs.DirEntry, error) {

	dir, err := os.ReadDir(path)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	var (
		files []fs.DirEntry
	)

	for _, v := range dir {
		if !v.IsDir() {
			files = append(files, v)
		}
	}

	return files, nil
}

func GetDirs(path string) ([]fs.DirEntry, error) {

	dir, err := os.ReadDir(path)
	if err != nil {
		return []fs.DirEntry{}, err
	}

	var (
		dirs []fs.DirEntry
	)

	for _, v := range dir {
		if v.IsDir() {
			dirs = append(dirs, v)
		}
	}

	return dirs, nil
}
