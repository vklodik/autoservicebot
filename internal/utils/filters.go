package utils

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"strconv"
)

func NoCommands(msg *gotgbot.Message) bool {
	return message.Text(msg) && !message.Command(msg)
}

func Confirms(cq *gotgbot.CallbackQuery) bool {
	return cq.Data == "yes" || cq.Data == "no"
}

func DateSelection(cq *gotgbot.CallbackQuery) bool {
	if cq.Data == PrevMonth || cq.Data == NextMonth {
		return true
	}

	dataInt, err := strconv.Atoi(cq.Data)
	if err != nil {
		return false
	}

	return dataInt >= 1 && dataInt <= 31
}

func TimeSelection(cq *gotgbot.CallbackQuery) bool {
	return cq.Data == "09:00" || cq.Data == "10:30" || cq.Data == "12:00" || cq.Data == "13:30" || cq.Data == "15:00" || cq.Data == "16:30" || cq.Data == "18:00" || cq.Data == "19:30"
}
