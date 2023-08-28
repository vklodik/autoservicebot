package sessions

import (
	"automobile36/internal/db"
	"automobile36/internal/utils"
	"fmt"
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers/filters/message"
	"os"
)

func LoadMenuHandlers(dp *ext.Dispatcher) {
	dp.AddHandler(handlers.NewMessage(message.Equal("Запись 📃"), SendRecordsMenu))
	dp.AddHandler(handlers.NewMessage(message.Equal("Прайс лист 💵"), SendPrice))
	dp.AddHandler(handlers.NewMessage(message.Equal("Наши контакты ☎"), SendContacts))
	dp.AddHandler(handlers.NewMessage(message.Equal("Мы на картах 🗺️"), SendLocation))
}

func SendRecordsMenu(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	name, number, err := db.GetInfo(int(ctx.EffectiveChat.Id))
	if err != nil {
		return fmt.Errorf("error while getting info about user: %w", err)
	}

	t := fmt.Sprintf(`<b>Ваши данные</b>
<b>Имя: %s</b>
<b>Номер телефона: %d</b>
Чтобы изменить номер, нажмите на кнопку "Изменить номер телефона".
Чтобы записаться нажмите "Добавить запись".`, name, number)
	if _, err := ctx.EffectiveChat.SendMessage(b, t, &gotgbot.SendMessageOpts{
		ParseMode:   "html",
		ReplyMarkup: utils.GetRecordsKeyboard(),
	}); err != nil {
		return fmt.Errorf("error while sending records menu: %w", err)
	}

	return nil
}

func SendPrice(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	file, err := os.Open("img/price.jpg")
	if err != nil {
		return fmt.Errorf("error while opening file: %w", err)
	}
	defer file.Close()

	// Отправляем фото
	_, err = b.SendPhoto(ctx.EffectiveChat.Id, file, &gotgbot.SendPhotoOpts{})
	if err != nil {
		return fmt.Errorf("error while sending photo: %w", err)
	}

	return nil
}

func SendContacts(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	_, err := ctx.EffectiveChat.SendMessage(b, "Номера телефонов:\n+7XXXXXXXXXX\n7XXXXXXXXXX\n\nМы ВК: https://vk.com/XXXXXXXXXXXX", nil)
	if err != nil {
		return fmt.Errorf("error while sending contacts: %w", err)
	}

	return nil
}

func SendLocation(b *gotgbot.Bot, ctx *ext.Context) error {
	if ctx.EffectiveChat.Type != "private" {
		return nil
	}
	_, err := b.SendLocation(ctx.EffectiveChat.Id, 55.754029, 37.620743, nil)
	if err != nil {
		return fmt.Errorf("error while sending location: %w", err)
	}

	return nil
}
