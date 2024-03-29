package likeforce

import (
	"errors"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ExtractChatID extracts ID of the chat from which update goes
func ExtractChatID(update tgbotapi.Update) (int64, error) {
	if update.Message != nil {
		return update.Message.Chat.ID, nil
	}
	if update.CallbackQuery != nil {
		if update.CallbackQuery.Message != nil {
			return update.CallbackQuery.Message.Chat.ID, nil
		}

		// if update.CallbackQuery.ChatInstance

		parts := strings.SplitN(update.CallbackQuery.Data, ":", 2)
		return strconv.ParseInt(parts[0], 10, 64)
	}
	return 0, errors.New("cannot extract chat id")
}

// ExtractPostID extracts ID of the message that button was reply on
func ExtractPostID(update tgbotapi.Update) (int, error) {
	if update.CallbackQuery == nil {
		return update.Message.MessageID, nil
	}

	if update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.ReplyToMessage.MessageID, nil
	}
	parts := strings.SplitN(update.CallbackQuery.Data, ":", 2)
	return strconv.Atoi(parts[1])
}

// ExtractButtonID extract ID of the message with button from callback request
func ExtractButtonID(update tgbotapi.Update) (int, error) {
	if update.CallbackQuery == nil {
		return 0, errors.New("can extract button ID only from callback response")
	}
	if update.CallbackQuery.Message != nil {
		return update.CallbackQuery.Message.MessageID, nil
	}
	postID, err := ExtractPostID(update)
	return postID + 1, err
}
