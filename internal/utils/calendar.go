package utils

import (
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/patrickmn/go-cache"
	"time"
)

type CalendarCallback struct {
	Year  int
	Month time.Month
	day   int
}

const (
	IGNORE    string = "nothing"
	PrevMonth string = "prev_month"
	NextMonth string = "next_month"
)

var weekDays = [7]string{"Пн", "Вт", "Ср", "Чт", "Пт", "Сб", "Вс"}
var monthNames = [12]string{"Январь", "Февраль", "Март", "Апрель", "Май", "Июнь", "Июль", "Август", "Сентябрь", "Октябрь", "Ноябрь", "Декабрь"}
var CalendarCache = cache.New(5*time.Minute, 10*time.Minute)

func SimpleCalendar(userId string, year int, month time.Month) gotgbot.InlineKeyboardMarkup {
	var kb [][]gotgbot.InlineKeyboardButton
	data := CalendarCallback{
		Year:  year,
		Month: month,
		day:   1,
	}
	CalendarCache.Set(userId, data, cache.DefaultExpiration)

	monthName := []gotgbot.InlineKeyboardButton{{Text: fmt.Sprintf("%s %d", monthNames[month-1], year), CallbackData: IGNORE}}
	kb = append(kb, monthName)

	var weekDaysRow []gotgbot.InlineKeyboardButton
	for _, day := range weekDays {
		weekDaysRow = append(weekDaysRow, gotgbot.InlineKeyboardButton{Text: day, CallbackData: IGNORE})
	}
	kb = append(kb, weekDaysRow)

	totalDays := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	initWeekDay := int(time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).Weekday())
	if initWeekDay == 0 {
		initWeekDay = 7 // Коррекция значения воскресенья на 7
	}

	calendar := [6][7]int{}
	for i := 1; i <= totalDays; i++ {
		row := (i - 1 + initWeekDay - 1) / 7
		col := (i - 1 + initWeekDay - 1) % 7
		calendar[row][col] = i
	}

	for _, week := range calendar {
		var row []gotgbot.InlineKeyboardButton
		for _, day := range week {
			if day != 0 {
				button := gotgbot.InlineKeyboardButton{Text: fmt.Sprintf("%d", day), CallbackData: fmt.Sprintf("%d", day)}
				row = append(row, button)
			} else {
				button := gotgbot.InlineKeyboardButton{Text: " ", CallbackData: IGNORE}
				row = append(row, button)
			}
		}

		isEmptyRow := true
		for _, button := range row {
			if button.Text != " " {
				isEmptyRow = false
				break
			}
		}

		if !isEmptyRow {
			kb = append(kb, row)
		}
	}

	selectMonthRow := []gotgbot.InlineKeyboardButton{
		{Text: "<", CallbackData: PrevMonth},
		{Text: ">", CallbackData: NextMonth},
	}
	kb = append(kb, selectMonthRow)

	return gotgbot.InlineKeyboardMarkup{InlineKeyboard: kb}
}
