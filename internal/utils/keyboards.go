package utils

import (
	"automobile36/internal/db"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"time"
)

func GetConfirmKeyboard() gotgbot.InlineKeyboardMarkup {
	b1 := gotgbot.InlineKeyboardButton{Text: "Да ✅", CallbackData: "yes"}
	b2 := gotgbot.InlineKeyboardButton{Text: "Нет ❌", CallbackData: "no"}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{b1, b2},
		},
	}
}

func GetMenuKeyboard() gotgbot.ReplyKeyboardMarkup {
	b1 := gotgbot.KeyboardButton{Text: "Запись 📃"}
	b2 := gotgbot.KeyboardButton{Text: "Прайс лист 💵"}
	b3 := gotgbot.KeyboardButton{Text: "Наши контакты ☎"}
	b4 := gotgbot.KeyboardButton{Text: "Мы на картах 🗺️"}

	return gotgbot.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]gotgbot.KeyboardButton{
			{b1, b2},
			{b3, b4},
		},
	}
}

func GetRecordsKeyboard() gotgbot.ReplyKeyboardMarkup {
	b1 := gotgbot.KeyboardButton{Text: "Добавить запись 📝"}
	b2 := gotgbot.KeyboardButton{Text: "Изменить номер телефона 📱"}
	b3 := gotgbot.KeyboardButton{Text: "Ваши записи 📜"}
	b4 := gotgbot.KeyboardButton{Text: "Назад 👈"}

	return gotgbot.ReplyKeyboardMarkup{
		ResizeKeyboard: true,
		Keyboard: [][]gotgbot.KeyboardButton{
			{b1, b2},
			{b3, b4},
		},
	}
}

func GetTimesKeyboard(result int64) (gotgbot.InlineKeyboardMarkup, error) {
	kb := [][]gotgbot.InlineKeyboardButton{{}}

	times, err := db.GetAllTimes(result)
	if err != nil {
		return gotgbot.InlineKeyboardMarkup{}, fmt.Errorf("error while getting times: %w", err)
	}

	var row []gotgbot.InlineKeyboardButton
	for _, t := range times {
		button := gotgbot.InlineKeyboardButton{Text: t, CallbackData: t}
		row = append(row, button)

		if len(row) == 2 {
			kb = append(kb, row)
			row = []gotgbot.InlineKeyboardButton{}
		}
	}

	if len(row) > 0 {
		kb = append(kb, row)
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: kb,
	}, nil
}

func GetAllUserRecordsKeyboard(records []int) gotgbot.InlineKeyboardMarkup {
	kb := [][]gotgbot.InlineKeyboardButton{{}}
	for _, record := range records {
		tm := time.Unix(int64(record), 0).Add(-3 * time.Hour)
		textTime := tm.Format("02.01.2006 15:04")
		kb = append(kb, []gotgbot.InlineKeyboardButton{{Text: textTime, CallbackData: IGNORE}})
	}

	return gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: kb,
	}
}
