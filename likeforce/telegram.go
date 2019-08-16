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
	users    Users
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

	err := tg.posts.Add(chatID, postID)
	if err != nil {
		tg.logger.ErrorWith("cannot add post").Err("error", err).Write()
		return
	}

	stat, err := tg.users.Stat(chatID, userID)
	if err != nil {
		tg.logger.ErrorWith("cannot get stat for user").Err("error", err).Write()
		return
	}

	err = tg.users.AddPost(chatID, userID)
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

func (tg *Telegram) processButton(update tgbotapi.Update) {
	msg := update.CallbackQuery
	tg.logger.InfoWith("new button request").String("from", msg.From.String()).Write()

	// parse IDs
	userID := msg.From.ID
	chatID, err := ExtractChatId(update)
	if err != nil {
		tg.logger.ErrorWith("cannot extract chat id").Err("error", err).Write()
		return
	}
	postID, err := ExtractPostId(update)
	if err != nil {
		tg.logger.ErrorWith("cannot extract post id").Err("error", err).Write()
		return
	}
	tg.logger.DebugWith("ids").Int64("chat", chatID).Int("post", postID).Int("user", userID).Write()

	// check post existence
	postExists, err := tg.posts.Has(chatID, postID)
	if err != nil {
		tg.logger.ErrorWith("cannot check post existence").Err("error", err).Write()
		return
	}
	if !postExists {
		tg.logger.WarnWith("cannot find post").Err("error", err).Write()
		_, err := tg.bot.AnswerCallbackQuery(
			tgbotapi.NewCallback(msg.ID, tg.messages.Error),
		)
		if err != nil {
			tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
		}
		return
	}

	// dislike post if laready liked, like otherwise
	liked, err := tg.likes.Has(chatID, postID, userID)
	if err != nil {
		tg.logger.ErrorWith("cannot check like existence").Err("error", err).Write()
		return
	}
	if liked {
		err = tg.likes.Remove(chatID, postID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot remove like").Err("error", err).Write()
			return
		}
		err = tg.users.RemoveRating(chatID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot decrement rating").Err("error", err).Write()
			return
		}
	} else {
		err = tg.likes.Add(chatID, postID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot add like").Err("error", err).Write()
			return
		}
		err = tg.users.AddRating(chatID, userID)
		if err != nil {
			tg.logger.ErrorWith("cannot increment rating").Err("error", err).Write()
			return
		}
	}

	// update counter on button
	buttonID, err := ExtractButtonID(update)
	if err != nil {
		tg.logger.ErrorWith("cannot get button ID").Err("error", err).Write()
		return
	}
	likesCount, err := tg.likes.Count(chatID, postID)
	_, err = tg.bot.Send(
		tgbotapi.NewEditMessageReplyMarkup(chatID, buttonID, tg.makeButton(chatID, buttonID, likesCount)),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot update button").Err("error", err).Write()
		return
	}

	// send response
	var responseText string
	if liked {
		responseText = tg.messages.Disliked
	} else {
		responseText = tg.messages.Liked
	}
	_, err = tg.bot.AnswerCallbackQuery(
		tgbotapi.NewCallback(msg.ID, responseText),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("button response sent").String("to", msg.From.String()).Write()
}

func (tg *Telegram) makeButton(chatID int64, postID int, likesCount int) tgbotapi.InlineKeyboardMarkup {
	var text string
	if likesCount == 0 {
		text = tg.messages.Like
	} else {
		text = fmt.Sprintf("%s - %d", tg.messages.Like, likesCount)
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(text, fmt.Sprintf("%d:%d", chatID, postID)),
		),
	)
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
func NewTelegram(config Config, likes Likes, posts Posts, users Users, logger *onelog.Logger) (Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(config.Telegram.Token)
	if err != nil {
		return Telegram{}, err
	}
	bot.Debug = config.Telegram.Debug
	tg := Telegram{
		likes:    likes,
		posts:    posts,
		users:    users,
		bot:      bot,
		timeout:  config.Telegram.Timeout,
		messages: config.Messages,
		logger:   logger,
	}
	return tg, nil
}
