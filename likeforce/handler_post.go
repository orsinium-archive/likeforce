package likeforce

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"

// PostHandler psorcesses a new post in a group
type PostHandler struct {
	*Telegram
}

// Match returns true if handler can process given Telegram update
func (tg *PostHandler) Match(update tgbotapi.Update) bool {
	if update.Message == nil {
		return false
	}
	if update.Message.Chat.Type == "private" {
		return false
	}

	if update.Message.Text != "" {
		return true
	}
	if update.Message.Caption != "" {
		return true
	}

	return false
}

// Handle do actions for given Telegram update
func (tg *PostHandler) Handle(update tgbotapi.Update) {
	tg.logger.InfoWith("new message").String("from", update.Message.From.String()).Write()

	chat := tg.storage.Chat(update.Message.Chat.ID)
	post := chat.Post(update.Message.MessageID)
	user := chat.User(update.Message.From.ID)

	err := post.Author(user.ID).Create()
	if err != nil {
		tg.logger.ErrorWith("cannot add post").Err("error", err).Write()
		return
	}

	stat, err := UserStat(*user)
	if err != nil {
		tg.logger.ErrorWith("cannot get stat for user").Err("error", err).Write()
		return
	}

	err = user.SetName(update.Message.From.String())
	if err != nil {
		tg.logger.ErrorWith("cannot save username").Err("error", err).Write()
		return
	}

	msg := tgbotapi.NewMessage(chat.ID, stat)
	msg.ReplyToMessageID = post.ID
	msg.DisableNotification = true
	msg.ParseMode = "Markdown"
	msg.ReplyMarkup = tg.makeButton(chat.ID, post.ID, 0)
	_, err = tg.bot.Send(msg)
	if err != nil {
		tg.logger.ErrorWith("cannot send message").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("message sent").String("to", update.Message.From.String()).Write()
}
