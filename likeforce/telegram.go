package likeforce

import (
	"fmt"

	"github.com/francoispqt/onelog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Telegram is a main logic for handling messages
type Telegram struct {
	likes    Likes
	posts    Posts
	bot      *tgbotapi.BotAPI
	timeout  int
	messages MessagesConfig
	logger   *onelog.Logger
}

func (tg *Telegram) processMessage(update tgbotapi.Update) {
	tg.logger.InfoWith("new message").String("from", update.Message.From.String()).Write()

	err := tg.posts.Add(update.Message.Chat.ID, update.Message.MessageID)
	if err != nil {
		tg.logger.ErrorWith("cannot add post").Err("error", err).Write()
		return
	}

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	msg.ReplyToMessageID = update.Message.MessageID
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(tg.messages.Like, fmt.Sprintf("%d", update.Message.MessageID)),
		),
	)
	_, err = tg.bot.Send(msg)
	if err != nil {
		tg.logger.ErrorWith("cannot send message").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("message sent").String("to", update.Message.From.String()).Write()
}

func (tg *Telegram) processButton(update tgbotapi.Update) {
	tg.logger.InfoWith("new button request").String("from", update.CallbackQuery.From.String()).Write()

	_, err := tg.bot.AnswerCallbackQuery(
		tgbotapi.NewCallback(update.CallbackQuery.ID, tg.messages.Liked),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("button response sent").String("to", update.CallbackQuery.From.String()).Write()
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
func NewTelegram(config Config, likes Likes, posts Posts, logger *onelog.Logger) (Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return Telegram{}, err
	}
	bot.Debug = config.Telegram.Debug
	tg := Telegram{
		likes:    likes,
		posts:    posts,
		bot:      bot,
		timeout:  config.Telegram.Timeout,
		messages: config.Messages,
		logger:   logger,
	}
	return tg, nil
}
