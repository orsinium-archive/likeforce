package likeforce

import (
	"fmt"
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Telegram is a main logic for handling messages
type Telegram struct {
	likes    Likes
	bot      *tgbotapi.BotAPI
	timeout  int
	messages MessagesConfig
}

func (tg *Telegram) processMessage(update tgbotapi.Update) {
	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(tg.messages.Like, fmt.Sprintf("%d", update.Message.MessageID)),
		),
	)
	tg.bot.Send(msg)
}

func (tg *Telegram) processButton(update tgbotapi.Update) {
	fmt.Println(update.CallbackQuery.From.UserName)
	fmt.Println(update.CallbackQuery.Data)

	tg.bot.AnswerCallbackQuery(tgbotapi.NewCallback(update.CallbackQuery.ID, tg.messages.Liked))
}

// Serve forever to process all incoming messages
func (tg *Telegram) Serve() error {
	u := tgbotapi.NewUpdate(0)
	if tg.timeout != 0 {
		u.Timeout = tg.timeout
	}

	updates, err := tg.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		if update.CallbackQuery != nil {
			tg.processButton(update)
		}
		if update.Message != nil { // ignore any non-Message Updates
			tg.processMessage(update)
		}
	}
	return nil
}

// NewTelegram creates Telegram instance
func NewTelegram(config Config, likes Likes) (Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return Telegram{}, err
	}
	bot.Debug = config.Telegram.Debug
	tg := Telegram{
		likes:    likes,
		bot:      bot,
		timeout:  config.Telegram.Timeout,
		messages: config.Messages,
	}
	return tg, nil
}
