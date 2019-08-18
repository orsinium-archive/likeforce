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
}

func (tg *Telegram) processMessage(update tgbotapi.Update) {
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
	msg.ReplyMarkup = tg.makeButton(chat.ID, post.ID, 0)
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

func (tg *Telegram) processDigest(update tgbotapi.Update) string {
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

	if update.Message != nil && update.Message.Chat.Type != "private" {
		if update.Message.Text == "/digest" && update.Message.From.UserName == tg.admin {
			// process the digest request
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, tg.processDigest(update))
			msg.ParseMode = "Markdown"
			_, err := tg.bot.Send(msg)
			if err != nil {
				tg.logger.ErrorWith("cannot send message").Err("error", err).Write()
				return
			}
			tg.logger.InfoWith("message sent").String("to", update.Message.From.String()).Write()
		} else {
			// process a new post in the group
			tg.processMessage(update)
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
func NewTelegram(config Config, storage storage.Storage, logger *onelog.Logger) (Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return Telegram{}, err
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
	return tg, nil
}
