package likeforce

import (
	"fmt"

	"github.com/francoispqt/onelog"
	"github.com/orsinium/likeforce/likeforce/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Telegram is a main logic for handling messages
type Telegram struct {
	storage  storage.Storage
	bot      *tgbotapi.BotAPI
	timeout  int
	admin    string
	messages MessagesConfig
	logger   *onelog.Logger

	handlers []Handler
}

// RegisterHandler adds new Handler for processing of Telegram updates
func (tg *Telegram) RegisterHandler(handler Handler) {
	tg.handlers = append(tg.handlers, handler)
}

func (tg *Telegram) makeButton(chatID int64, postID int, likesCount int) tgbotapi.InlineKeyboardMarkup {
	text := getButtonText(likesCount, tg.messages.Like)
	if likesCount > 0 {
		text = fmt.Sprintf("%s %s", text, ByteCount(likesCount))
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, fmt.Sprintf("%d:%d", chatID, postID)),
		),
	)
}

func getButtonText(likesCount int, messages []string) string {
	if len(messages) == 1 || likesCount == 0 {
		return messages[0]
	}
	index := likesCount/10 + 1
	if index >= len(messages) {
		return messages[len(messages)-1]
	}
	return messages[index]
}

func (tg *Telegram) processUpdate(update tgbotapi.Update) {
	for _, handler := range tg.handlers {
		if handler.Match(update) {
			handler.Handle(update)
			return
		}
	}
}

// Serve forever to process all incoming messages
func (tg *Telegram) Serve() error {
	tg.logger.Info("serve")
	u := tgbotapi.NewUpdate(0)
	if tg.timeout != 0 {
		u.Timeout = tg.timeout
	}

	updates, err := tg.bot.GetUpdatesChan(u)
	if err != nil {
		return err
	}

	for update := range updates {
		tg.logger.Debug("new update")
		go tg.processUpdate(update)
	}
	return nil
}

// NewTelegram creates Telegram instance
func NewTelegram(config Config, storage storage.Storage, logger *onelog.Logger) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return &Telegram{}, err
	}
	bot.Debug = config.Telegram.Debug
	tg := Telegram{
		storage:  storage,
		bot:      bot,
		timeout:  config.Telegram.Timeout,
		admin:    config.Telegram.Admin,
		messages: config.Messages,
		logger:   logger,
	}
	return &tg, nil
}
