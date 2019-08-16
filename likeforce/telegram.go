package likeforce

import (
	"fmt"
	"strconv"
	"strings"

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
			tgbotapi.NewInlineKeyboardButtonData(
				tg.messages.Like,
				fmt.Sprintf("%d:%d", update.Message.Chat.ID, update.Message.MessageID)),
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
	msg := update.CallbackQuery
	tg.logger.InfoWith("new button request").String("from", msg.From.String()).Write()

	// parse chat and post IDs
	parts := strings.SplitN(msg.Data, ":", 2)
	chatID, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		tg.logger.ErrorWith("cannot extract chat id").Err("error", err).Write()
		return
	}
	postID, err := strconv.Atoi(parts[1])
	if err != nil {
		tg.logger.ErrorWith("cannot extract post id").Err("error", err).Write()
		return
	}

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
	liked, err := tg.likes.Has(chatID, postID, msg.From.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot check like existence").Err("error", err).Write()
		return
	}
	var responseText string
	if liked {
		responseText = tg.removeLike(chatID, postID, msg)
	} else {
		responseText = tg.addLike(chatID, postID, msg)
	}

	// send response
	_, err = tg.bot.AnswerCallbackQuery(
		tgbotapi.NewCallback(msg.ID, responseText),
	)
	if err != nil {
		tg.logger.ErrorWith("cannot send callback answer").Err("error", err).Write()
		return
	}
	tg.logger.InfoWith("button response sent").String("to", msg.From.String()).Write()
}

func (tg *Telegram) addLike(chatID int64, postID int, msg *tgbotapi.CallbackQuery) string {
	err := tg.likes.Add(chatID, postID, msg.From.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot add like").Err("error", err).Write()
		return tg.messages.Error
	}
	return tg.messages.Liked
}

func (tg *Telegram) removeLike(chatID int64, postID int, msg *tgbotapi.CallbackQuery) string {
	err := tg.likes.Remove(chatID, postID, msg.From.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot remove like").Err("error", err).Write()
		return tg.messages.Error
	}
	return tg.messages.Disliked
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
