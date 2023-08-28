package sessions

import (
	"automobile36/internal/db"
	"automobile36/internal/utils"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"github.com/patrickmn/go-cache"
	"strconv"
	"time"
)

const (
	SELECT = "select"
	TIME   = "time"
	CHANGE = "change"
)

var recordsCache = cache.New(5*time.Minute, 10*time.Minute)

func LoadRecordsHandlers(dp *ext.Dispatcher) {
	dp.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewMessage(message.Equal("Добавить запись 📝"), AddNewRecord)},
		map[string][]ext.Handler{
			SELECT:  {handlers.NewCallback(utils.DateSelection, ProcessSelection)},
			TIME:    {handlers.NewCallback(utils.TimeSelection, SelectTime)},
			CONFIRM: {handlers.NewCallback(utils.Confirms, ConfirmRecord)},
		},
		&handlers.ConversationOpts{
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		},
	))
	dp.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewMessage(message.Equal("Изменить номер телефона 📱"), ChangePhoneNumber)},
		map[string][]ext.Handler{
			CHANGE:  {handlers.NewMessage(utils.NoCommands, AddNewNumber)},
			CONFIRM: {handlers.NewCallback(utils.Confirms, ConfirmNewPhoneNumber)},
		},
		&handlers.ConversationOpts{
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		},
	))

	dp.AddHandler(handlers.NewMessage(message.Equal("Ваши записи 📜"), ListAllRecords))
	dp.AddHandler(handlers.NewMessage(message.Equal("Назад 👈"), GoBack))
}

func AddNewRecord(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	msg, err := b.SendMessage(ctx.EffectiveChat.Id, ".", &gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true}})
	if err != nil {
		return fmt.Errorf("error while deleting keyboard: %w", err)
	}

	_, err = b.DeleteMessage(ctx.EffectiveChat.Id, msg.MessageId, &gotgbot.DeleteMessageOpts{})
	if err != nil {
		return fmt.Errorf("error while deleting message: %w", err)
	}

	if _, err := ctx.EffectiveChat.SendMessage(
		b,
		"Выберите дату",
		&gotgbot.SendMessageOpts{
			ReplyMarkup: utils.SimpleCalendar(strconv.Itoa(int(ctx.EffectiveChat.Id)), time.Now().Year(), time.Now().Month()),
		}); err != nil {
		return fmt.Errorf("error while sending calendar: %w", err)
	}

	return handlers.NextConversationState(SELECT)
}

func ProcessSelection(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	chatId := strconv.Itoa(int(ctx.EffectiveChat.Id))
	data, foundData := utils.CalendarCache.Get(chatId)
	if !foundData {
		return fmt.Errorf("failed to get data from cache")
	}
	newData := data.(utils.CalendarCallback)
	tempTime := time.Date(newData.Year, newData.Month, 1, 0, 0, 0, 0, time.UTC)

	switch cb.Data {
	case utils.IGNORE:
	case utils.PrevMonth:
		if tempTime.Month() != time.Now().Month() {
			prevDate := tempTime.Add(-24 * time.Hour)
			_, _, err := ctx.EffectiveMessage.EditReplyMarkup(
				b,
				&gotgbot.EditMessageReplyMarkupOpts{
					ReplyMarkup: utils.SimpleCalendar(chatId, prevDate.Year(), prevDate.Month()),
				})
			if err != nil {
				return fmt.Errorf("failed to edit markup")
			}
		}
	case utils.NextMonth:
		nextDate := tempTime.AddDate(0, 1, 0)
		_, _, err := ctx.EffectiveMessage.EditReplyMarkup(
			b,
			&gotgbot.EditMessageReplyMarkupOpts{
				ReplyMarkup: utils.SimpleCalendar(chatId, nextDate.Year(), nextDate.Month()),
			})
		if err != nil {
			return fmt.Errorf("failed to edit markup")
		}
	default:
		newDay := cb.Data
		newDayInt, err := strconv.Atoi(newDay)
		if err != nil {
			return fmt.Errorf("failed to convert string to int")
		}

		result := time.Date(newData.Year, newData.Month, newDayInt, 0, 0, 0, 0, time.UTC)
		compTime := time.Now().AddDate(0, 0, -1)
		if result.Before(compTime) {
			_, _, err := ctx.EffectiveMessage.EditText(
				b,
				fmt.Sprintf("Нельзя выбрать: %s (прошедшую дату)!\nПопробуйте снова", result.Format("02.01.2006")),
				&gotgbot.EditMessageTextOpts{
					ReplyMarkup: utils.SimpleCalendar(strconv.Itoa(int(ctx.EffectiveChat.Id)), time.Now().Year(), time.Now().Month()),
				})
			if err != nil {
				return fmt.Errorf("error while sending calendar: %w", err)
			}

			return handlers.NextConversationState(SELECT)
		}

		if _, err := ctx.EffectiveMessage.Delete(b, &gotgbot.DeleteMessageOpts{}); err != nil {
			return fmt.Errorf("failed to delete message")
		}

		recordsCache.Set(strconv.FormatInt(ctx.EffectiveChat.Id, 10)+"_chosen_date", result, cache.DefaultExpiration)
		kb, err := utils.GetTimesKeyboard(result.Unix())
		if err != nil {
			return fmt.Errorf("error while getting times kb: %w", err)
		}

		if _, err := ctx.EffectiveChat.SendMessage(
			b,
			fmt.Sprintf("Вы выбрали: %s", result.Format(time.DateOnly)),
			&gotgbot.SendMessageOpts{ReplyMarkup: kb},
		); err != nil {
			return fmt.Errorf("error while sending date: %w", err)
		}

		return handlers.NextConversationState(TIME)
	}

	return nil
}

func SelectTime(b *gotgbot.Bot, ctx *ext.Context) error {
	chosenDateInterface, ok := recordsCache.Get(strconv.FormatInt(ctx.EffectiveChat.Id, 10) + "_chosen_date")
	if !ok {
		return fmt.Errorf("error while getting date from cache")
	}

	chosenDate := chosenDateInterface.(time.Time)
	cb := ctx.Update.CallbackQuery

	parsedTime, err := time.Parse("15:04", cb.Data)
	if err != nil {
		return fmt.Errorf("error while ...: %w", err)
	}

	sum := chosenDate.Add(time.Duration(parsedTime.Hour()) * time.Hour)
	sum = sum.Add(time.Duration(parsedTime.Minute()) * time.Minute)
	recordsCache.Set(strconv.FormatInt(ctx.EffectiveChat.Id, 10)+"_datetime", sum.Unix(), cache.DefaultExpiration)

	if _, _, err := ctx.EffectiveMessage.EditText(
		b,
		fmt.Sprintf("Дата: %s\nВремя: %s", chosenDate.Format("02.01.2006"), cb.Data),
		&gotgbot.EditMessageTextOpts{ReplyMarkup: utils.GetConfirmKeyboard()},
	); err != nil {
		return fmt.Errorf("error while ...: %w", err)
	}

	return handlers.NextConversationState(CONFIRM)
}

func ConfirmRecord(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery
	switch cb.Data {
	case "yes":
		unixDatetime, ok := recordsCache.Get(strconv.FormatInt(ctx.EffectiveChat.Id, 10) + "_datetime")
		if !ok {
			return fmt.Errorf("error while confirming record")
		}
		err := db.SaveRecord(ctx.EffectiveChat.Id, unixDatetime.(int64))
		if err != nil {
			return fmt.Errorf("error while saving record: %w", err)
		}
		if _, _, err := ctx.EffectiveMessage.EditText(b, "Вы успешно записались!", nil); err != nil {
			return fmt.Errorf("error while confirming record: %w", err)
		}

		name, number, err := db.GetInfo(int(ctx.EffectiveChat.Id))
		if err != nil {
			return fmt.Errorf("error while getting info about user: %w", err)
		}
		t := time.Unix(unixDatetime.(int64)-3*60*60, 0).Format("02.01.2006 15:04")
		// prod: -1001891091220			test: -673660970
		recordsChat := gotgbot.Chat{Id: -1001891091220, Type: "group"}
		if _, err := recordsChat.SendMessage(
			b,
			fmt.Sprintf("Запись на %s\nИмя клиента: %s\nНомер телефона: %d", t, name, number),
			&gotgbot.SendMessageOpts{ReplyMarkup: utils.GetRecordsKeyboard()},
		); err != nil {
			return fmt.Errorf("error while back up to menu: %w", err)
		}
		if _, err := ctx.EffectiveChat.SendMessage(
			b,
			"Возвращаемся в меню",
			&gotgbot.SendMessageOpts{ReplyMarkup: utils.GetRecordsKeyboard()},
		); err != nil {
			return fmt.Errorf("error while back up to menu: %w", err)
		}

		return handlers.EndConversation()
	case "no":
		_, _, err := ctx.EffectiveMessage.EditText(
			b,
			"Попробуем снова!\nВыберите дату",
			&gotgbot.EditMessageTextOpts{
				ReplyMarkup: utils.SimpleCalendar(strconv.Itoa(int(ctx.EffectiveChat.Id)), time.Now().Year(), time.Now().Month()),
			})
		if err != nil {
			return fmt.Errorf("error while sending calendar: %w", err)
		}

		return handlers.NextConversationState(SELECT)
	}

	return nil
}

func ChangePhoneNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	_, err := ctx.EffectiveChat.SendMessage(
		b,
		"Отправьте новый номер телефона",
		&gotgbot.SendMessageOpts{ReplyMarkup: gotgbot.ReplyKeyboardRemove{RemoveKeyboard: true}},
	)

	if err != nil {
		return fmt.Errorf("error while asking for a new number: %w", err)
	}

	return handlers.NextConversationState(CHANGE)
}

func AddNewNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	inputNumber := ctx.EffectiveMessage.Text

	if _, err := strconv.Atoi(inputNumber); err != nil {
		_, err := ctx.EffectiveChat.SendMessage(b, "Номер должен состоять из цифр!\nпопробуйте ещё раз", nil)
		if err != nil {
			return fmt.Errorf("error while sending number check message: %w", err)
		}
		return nil
	}

	recordsCache.Set(strconv.FormatInt(ctx.EffectiveChat.Id, 10)+"_upd_number", inputNumber, cache.DefaultExpiration)

	_, err := ctx.EffectiveChat.SendMessage(
		b,
		fmt.Sprintf("Новый номер: %s\nПодтвердить?", inputNumber),
		&gotgbot.SendMessageOpts{
			ReplyMarkup: utils.GetConfirmKeyboard(),
		},
	)

	if err != nil {
		return fmt.Errorf("error while asking for number confirmation: %w", err)
	}

	return handlers.NextConversationState(CONFIRM)
}

func ConfirmNewPhoneNumber(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	switch cb.Data {
	case "yes":
		chatID := strconv.FormatInt(ctx.EffectiveChat.Id, 10)
		number, foundNumber := recordsCache.Get(chatID + "_upd_number")

		if !foundNumber {
			// Если данные не найдены в кэше, обработка ошибки
			return fmt.Errorf("failed to get name and/or number from cache")
		}

		numberStr, ok := number.(string)
		if !ok {
			return fmt.Errorf("failed to convert number to string")
		}

		numberInt, err := strconv.Atoi(numberStr)
		if err != nil {
			return fmt.Errorf("failed to convert number to int: %w", err)
		}

		if err := db.UpdateNumber(numberInt, int(ctx.EffectiveChat.Id)); err != nil {
			return err
		}

		if _, err := ctx.EffectiveMessage.Delete(b, nil); err != nil {
			return fmt.Errorf("failed to delete message: %w", err)
		}
		if _, err := ctx.EffectiveChat.SendMessage(b, "Данные успешно сохранены!\nДобро пожаловать в главное меню!", &gotgbot.SendMessageOpts{ReplyMarkup: utils.GetMenuKeyboard()}); err != nil {
			return fmt.Errorf("failed to send success message: %w", err)
		}
		return handlers.EndConversation()
	case "no":
		_, _, err := ctx.EffectiveMessage.EditText(b, "Давайте начнём сначала!\nОтправьте новый номер телефона", nil)
		if err != nil {
			return fmt.Errorf("failed to send reset message: %w", err)
		}
		return handlers.NextConversationState(CHANGE)
	}

	return nil
}

func ListAllRecords(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	records, err := db.GetAllRecords(ctx.EffectiveChat.Id)
	if err != nil {
		return fmt.Errorf("error while getting all records: %w", err)
	}
	if len(records) > 0 {
		if _, err := ctx.EffectiveChat.SendMessage(b, "Ваши актуальные записи", &gotgbot.SendMessageOpts{ReplyMarkup: utils.GetAllUserRecordsKeyboard(records)}); err != nil {
			return fmt.Errorf("error while listing all records: %w", err)
		}
	} else {
		if _, err := ctx.EffectiveChat.SendMessage(b, "У вас нет актуальных записей", nil); err != nil {
			return fmt.Errorf("error while listing all records: %w", err)
		}
	}

	return nil
}

func GoBack(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	_, err := ctx.EffectiveChat.SendMessage(b, "Добро пожаловать в меню!", &gotgbot.SendMessageOpts{ReplyMarkup: utils.GetMenuKeyboard()})
	if err != nil {
		return fmt.Errorf("error while going back: %w", err)
	}

	return nil
}
