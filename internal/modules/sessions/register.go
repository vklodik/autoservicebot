package sessions

import (
	"automobile36/internal/db"
	"automobile36/internal/utils"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/conversation"
	"github.com/patrickmn/go-cache"
	"html"
	"strconv"
	"time"
)

const (
	NAME    = "name"
	NUMBER  = "number"
	CONFIRM = "confirm"
)

var registrationCache = cache.New(5*time.Minute, 10*time.Minute)

func LoadRegisterHandlers(dp *ext.Dispatcher) {
	dp.AddHandler(handlers.NewConversation(
		[]ext.Handler{handlers.NewCommand("start", Start)},
		map[string][]ext.Handler{
			NAME:    {handlers.NewMessage(utils.NoCommands, Name)},
			NUMBER:  {handlers.NewMessage(utils.NoCommands, Number)},
			CONFIRM: {handlers.NewCallback(utils.Confirms, ConfirmData)},
		},
		&handlers.ConversationOpts{
			StateStorage: conversation.NewInMemoryStorage(conversation.KeyStrategySenderAndChat),
		},
	))
}

// Start introduces the bot and starts the conversation
func Start(b *gotgbot.Bot, ctx *ext.Context) error {
	res, err := db.IsExists(int(ctx.EffectiveChat.Id))
	if err != nil {
		return fmt.Errorf("error while IsExists checks user: %w", err)
	}
	switch res {
	case true:
		if _, err := ctx.EffectiveChat.SendMessage(b, "Добро пожаловать в меню!", &gotgbot.SendMessageOpts{ReplyMarkup: utils.GetMenuKeyboard()}); err != nil {
			return fmt.Errorf("failed to send welcome message: %w", err)
		}
	case false:
		if _, err := ctx.EffectiveChat.SendMessage(b, "Привет, напишите как вас зовут?", &gotgbot.SendMessageOpts{}); err != nil {
			return fmt.Errorf("failed to send welcome message: %w", err)
		}

		return handlers.NextConversationState(NAME)
	}
	return nil
}

func Name(b *gotgbot.Bot, ctx *ext.Context) error {
	inputName := ctx.EffectiveMessage.Text

	// Сохраняем имя пользователя в кэше
	registrationCache.Set(strconv.FormatInt(ctx.EffectiveChat.Id, 10)+"_name", inputName, cache.DefaultExpiration)

	_, err := ctx.EffectiveMessage.Reply(
		b,
		fmt.Sprintf("Приятно познакомиться, %s!\n\nТеперь напишите ваш номер телефона.", html.EscapeString(inputName)),
		&gotgbot.SendMessageOpts{
			ParseMode: "html",
		})
	if err != nil {
		return fmt.Errorf("failed to send name message: %w", err)
	}
	return handlers.NextConversationState(NUMBER)
}

func Number(b *gotgbot.Bot, ctx *ext.Context) error {
	inputNumber := ctx.EffectiveMessage.Text

	if _, err := strconv.Atoi(inputNumber); err != nil {
		_, err := ctx.EffectiveChat.SendMessage(b, "Номер должен состоять из цифр!\nпопробуйте ещё раз", nil)
		if err != nil {
			return fmt.Errorf("error while sending number check message: %w", err)
		}
		return nil
	}

	name, found := registrationCache.Get(strconv.FormatInt(ctx.EffectiveChat.Id, 10) + "_name")
	if !found {
		// Если имя пользователя не найдено в кэше, обработка ошибки
		return fmt.Errorf("failed to get name from cache")
	}

	registrationCache.Set(strconv.FormatInt(ctx.EffectiveChat.Id, 10)+"_number", inputNumber, cache.DefaultExpiration)

	_, err := ctx.EffectiveChat.SendMessage(
		b,
		fmt.Sprintf("Имя: %s\nНомер телефона: %s\nВсё верно?", name.(string), inputNumber),
		&gotgbot.SendMessageOpts{
			ParseMode:   "html",
			ReplyMarkup: utils.GetConfirmKeyboard(),
		})
	if err != nil {
		return fmt.Errorf("failed to send number message: %w", err)
	}
	return handlers.NextConversationState(CONFIRM)
}

func ConfirmData(b *gotgbot.Bot, ctx *ext.Context) error {
	cb := ctx.Update.CallbackQuery

	switch cb.Data {
	case "yes":
		chatID := strconv.FormatInt(ctx.EffectiveChat.Id, 10)
		name, foundName := registrationCache.Get(chatID + "_name")
		number, foundNumber := registrationCache.Get(chatID + "_number")

		if !foundName || !foundNumber {
			// Если данные не найдены в кэше, обработка ошибки
			return fmt.Errorf("failed to get name and/or number from cache")
		}

		nameStr, ok := name.(string)
		if !ok {
			return fmt.Errorf("failed to convert name to string")
		}

		numberStr, ok := number.(string)
		if !ok {
			return fmt.Errorf("failed to convert number to string")
		}

		numberInt, err := strconv.Atoi(numberStr)
		if err != nil {
			return fmt.Errorf("failed to convert number to int: %w", err)
		}

		if err := db.SaveUser(int(ctx.EffectiveChat.Id), nameStr, numberInt); err != nil {
			return err
		}

		if _, err := ctx.EffectiveMessage.Delete(b, nil); err != nil {
			return fmt.Errorf("failed to send success message: %w", err)
		}
		if _, err := ctx.EffectiveChat.SendMessage(b, "Данные успешно сохранены!\nДобро пожаловать в главное меню!", &gotgbot.SendMessageOpts{ReplyMarkup: utils.GetMenuKeyboard()}); err != nil {
			return fmt.Errorf("failed to send welcome message: %w", err)
		}
		return handlers.EndConversation()
	case "no":
		_, _, err := ctx.EffectiveMessage.EditText(b, "Давайте начнём сначала!\nНапишите ваше имя", nil)
		if err != nil {
			return fmt.Errorf("failed to send reset message: %w", err)
		}
		return handlers.NextConversationState(NAME)
	}
	return nil
}
