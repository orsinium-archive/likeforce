package likeforce

import (
	"fmt"

	"github.com/francoispqt/onelog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Telegram is a main logic for handling messages
type Telegram struct {
	storage  Storage
	bot      *tgbotapi.BotAPI
	timeout  int
	messages MessagesConfig
	logger   *onelog.Logger
}

func (tg *Telegram) processMessage(update tgbotapi.Update) {
	tg.logger.InfoWith("new message").String("from", update.Message.From.String()).Write()
	chatID := update.Message.Chat.ID
	postID := update.Message.MessageID
	userID := update.Message.From.ID

	err := tg.storage.Posts.Add(chatID, postID)
	if err != nil {
		tg.logger.ErrorWith("cannot add post").Err("error", err).Write()
		return
	}

	stat, err := tg.storage.Users.Stat(chatID, userID)
	if err != nil {
		tg.logger.ErrorWith("cannot get stat for user").Err("error", err).Write()
		return
	}

	err = tg.storage.Users.AddPost(chatID, userID)
	if err != nil {
		tg.logger.ErrorWith("cannot increment posts for user").Err("error", err).Write()
		return
	}

	msg := tgbotapi.NewMessage(chatID, stat)
	msg.ReplyToMessageID = postID
	msg.ReplyMarkup = tg.makeButton(chatID, postID, 0)
	_, err = tg.bot.Send(msg)
	if err != nil {
		tg.logger.ErrorWith("cannot send message").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("message sent").String("to", update.Message.From.String()).Write()
}

func (tg *Telegram) processButton(update tgbotapi.Update) string {
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

	// check post existence
	postExists, err := tg.storage.Posts.Has(chatID, postID)
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

	// dislike post if laready liked, like otherwise
	liked, err := tg.storage.Likes.Has(chatID, postID, userID)
	if err != nil {
		tg.logger.ErrorWith("cannot check like existence").Err("error", err).Write()
		return tg.messages.Error
	}
	if liked {
		err = tg.storage.Likes.Remove(chatID, postID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot remove like").Err("error", err).Write()
			return tg.messages.Error
		}
		err = tg.storage.Users.RemoveRating(chatID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot decrement rating").Err("error", err).Write()
			return tg.messages.Error
		}
	} else {
		err = tg.storage.Likes.Add(chatID, postID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot add like").Err("error", err).Write()
			return tg.messages.Error
		}
		err = tg.storage.Users.AddRating(chatID, userID)
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
	likesCount, err := tg.storage.Likes.Count(chatID, postID)
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

func (tg *Telegram) makeButton(chatID int64, postID int, likesCount int) tgbotapi.InlineKeyboardMarkup {
	var text string
	if likesCount == 0 {
		text = tg.messages.Like
	} else {
		text = fmt.Sprintf("%s %s", tg.messages.Like, ByteCount(likesCount))
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, fmt.Sprintf("%d:%d", chatID, postID)),
		),
	)
}

func (tg *Telegram) processUpdate(update tgbotapi.Update) {
	// process pressed button
	if update.CallbackQuery != nil {
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

	// process a new message from group
	if update.Message != nil {
		tg.processMessage(update)
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
func NewTelegram(config Config, storage Storage, logger *onelog.Logger) (Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return Telegram{}, err
	}
	bot.Debug = config.Telegram.Debug
	tg := Telegram{
		storage:  storage,
		bot:      bot,
		timeout:  config.Telegram.Timeout,
		messages: config.Messages,
		logger:   logger,
	}
	return tg, nil
}
