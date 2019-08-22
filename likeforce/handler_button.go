package likeforce

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// ButtonHandler psorcesses a new post in a group
type ButtonHandler struct {
	*Telegram
}

// Match returns true if handler can process given Telegram update
func (tg *ButtonHandler) Match(update tgbotapi.Update) bool {
	return update.CallbackQuery != nil
}

// Handle do actions for given Telegram update
func (tg *ButtonHandler) Handle(update tgbotapi.Update) {
	responseText := tg.processButton(update)
	_, err := tg.bot.AnswerCallbackQuery(
		tgbotapi.NewCallback(update.CallbackQuery.ID, responseText),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
	} else {
		tg.logger.InfoWith("button response sent").String("to", update.CallbackQuery.From.String()).Write()
	}
}

func (tg *ButtonHandler) processButton(update tgbotapi.Update) string {
	msg := update.CallbackQuery
	tg.logger.InfoWith("new button request").String("from", msg.From.String()).Write()

	// parse IDs
	userID := msg.From.ID
	chatID, err := ExtractChatID(update)
	if err != nil {
		tg.logger.ErrorWith("cannot extract chat id").Err("error", err).Write()
		return tg.messages.Error
	}
	postID, err := ExtractPostID(update)
	if err != nil {
		tg.logger.ErrorWith("cannot extract post id").Err("error", err).Write()
		return tg.messages.Error
	}
	tg.logger.DebugWith("ids").Int64("chat", chatID).Int("post", postID).Int("user", userID).Write()

	// create storage managers
	chat := tg.storage.Chat(chatID)
	post := chat.Post(postID)
	user := chat.User(userID)
	like := post.Like(user.ID)

	// check post existence
	postExists, err := post.Exists()
	if err != nil {
		tg.logger.ErrorWith("cannot check post existence").Err("error", err).Write()
		return tg.messages.Error
	}
	if !postExists {
		tg.logger.WarnWith("cannot find post").Err("error", err).Write()
		_, err := tg.bot.AnswerCallbackQuery(
			tgbotapi.NewCallback(msg.ID, tg.messages.Error),
		)
		if err != nil {
			tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
		}
		return tg.messages.Error
	}

	// get author
	authorID, err := post.AuthorID()
	if err != nil {
		tg.logger.ErrorWith("cannot get post author").Err("error", err).Write()
		return tg.messages.Error
	}
	rating := chat.User(authorID).Rating()

	// forbid self-like
	if authorID == user.ID {
		return tg.messages.Self
	}

	// dislike post if already liked, like otherwise
	liked, err := like.Exists(user.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot check like existence").Err("error", err).Write()
		return tg.messages.Error
	}
	if liked {
		err = like.Remove(user.ID)
		if err != nil {
			tg.logger.ErrorWith("cannot remove like").Err("error", err).Write()
			return tg.messages.Error
		}
		err = rating.Decr()
		if err != nil {
			tg.logger.ErrorWith("cannot decrement rating").Err("error", err).Write()
			return tg.messages.Error
		}
	} else {
		err = like.Create(user.ID)
		if err != nil {
			tg.logger.ErrorWith("cannot add like").Err("error", err).Write()
			return tg.messages.Error
		}
		err = rating.Incr()
		if err != nil {
			tg.logger.ErrorWith("cannot increment rating").Err("error", err).Write()
			return tg.messages.Error
		}
	}

	// update counter on button
	buttonID, err := ExtractButtonID(update)
	if err != nil {
		tg.logger.ErrorWith("cannot get button ID").Err("error", err).Write()
		return tg.messages.Error
	}
	likesCount, err := post.Likes()
	if err != nil {
		tg.logger.ErrorWith("cannot get likes count").Err("error", err).Write()
		return tg.messages.Error
	}
	_, err = tg.bot.Send(
		tgbotapi.NewEditMessageReplyMarkup(chatID, buttonID, tg.makeButton(chatID, buttonID, likesCount)),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot update button").Err("error", err).Write()
		return tg.messages.Error
	}

	// send response
	if liked {
		return tg.messages.Disliked
	}
	return tg.messages.Liked

}
