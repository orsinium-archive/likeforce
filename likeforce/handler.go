package likeforce

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// Handler is an interface for Telegram update processors
type Handler interface {
	Match(update tgbotapi.Update) bool
	Handle(update tgbotapi.Update)
}
