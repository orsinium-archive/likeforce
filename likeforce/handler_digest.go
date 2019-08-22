package likeforce

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// DigestHandler psorcesses a new post in a group
type DigestHandler struct {
	*Telegram
}

// Match returns true if handler can process given Telegram update
func (tg *DigestHandler) Match(update tgbotapi.Update) bool {
	if update.Message == nil {
		return false
	}
	if update.Message.Chat.Type == "private" {
		return false
	}

	return update.Message.Text == "/digest"
}

// Handle do actions for given Telegram update
func (tg *DigestHandler) Handle(update tgbotapi.Update) {
	tg.logger.InfoWith("new /digest request").String("from", update.Message.From.String()).Write()

	if update.Message.From.UserName == tg.admin {
		// process the digest request from admin
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, tg.processDigest(update))
		msg.ParseMode = "Markdown"
		_, err := tg.bot.Send(msg)
		if err != nil {
			tg.logger.ErrorWith("cannot send message").Err("error", err).Write()
			return
		}
		tg.logger.InfoWith("message sent").String("to", update.Message.From.String()).Write()
	}

	// remove "/digest" message (from admin or non-admin, doesn't matter)
	tg.bot.DeleteMessage(
		tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID),
	)
}

func (tg *DigestHandler) processDigest(update tgbotapi.Update) string {
	chat := tg.storage.Chat(update.Message.Chat.ID)
	digestUsers, err := MakeDigestUsers(chat)
	if err != nil {
		tg.logger.ErrorWith("cannot make users digest").Err("error", err).Write()
		return tg.messages.Error
	}
	digestPosts, err := MakeDigestPosts(chat, update.Message.Chat.UserName)
	if err != nil {
		tg.logger.ErrorWith("cannot make posts digest").Err("error", err).Write()
		return tg.messages.Error
	}
	return digestUsers + "\n\n" + digestPosts
}
