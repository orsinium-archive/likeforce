package likeforce

import (
	"fmt"
	"sort"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type userInfo struct {
	name   string
	id     int
	rating int
}

func (tg *Telegram) processDigest(update tgbotapi.Update) string {
	digest, err := MakeDigest(tg.storage, update.Message.Chat.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot make digest").Err("error", err).Write()
		return tg.messages.Error
	}
	return digest
}

// MakeDigest returns text digest for given chat
func MakeDigest(storage Storage, chatID int64) (string, error) {
	// get users and their rating
	ids, err := storage.Users.List(chatID)
	if err != nil {
		return "", err
	}
	users := make([]userInfo, len(ids))
	for _, userID := range ids {
		rating, err := storage.Users.RatingCount(chatID, userID)
		if err != nil {
			return "", err
		}
		user := userInfo{name: "", id: userID, rating: rating}
		users = append(users, user)
	}

	// sort users by rating
	sort.Slice(users, func(i, j int) bool {
		return users[i].rating > users[j].rating
	})

	// make top 5 users
	digest := "*Top 5 authors:*"
	for i, user := range users[:min(5, len(users))] {
		if user.rating == 0 {
			continue
		}
		user.name, err = storage.Users.GetName(user.id)
		if err != nil {
			return "", err
		}
		digest += fmt.Sprintf("\n%d. [%s](tg://user?id=%d): %s", i+1, user.name, user.id, ByteCount(user.rating))
	}
	return digest, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
