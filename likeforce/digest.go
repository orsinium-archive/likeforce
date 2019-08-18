package likeforce

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joomcode/errorx"
)

type userInfo struct {
	name   string
	id     int
	rating int
}

type postInfo struct {
	title  string
	id     int
	rating int
}

func (tg *Telegram) processDigest(update tgbotapi.Update) string {
	digestUsers, err := MakeDigestUsers(tg.storage, update.Message.Chat.ID)
	if err != nil {
		tg.logger.ErrorWith("cannot make users digest").Err("error", err).Write()
		return tg.messages.Error
	}
	digestPosts, err := MakeDigestPosts(tg.storage, update.Message.Chat.ID, update.Message.Chat.UserName)
	if err != nil {
		tg.logger.ErrorWith("cannot make posts digest").Err("error", err).Write()
		return tg.messages.Error
	}
	return digestUsers + "\n\n" + digestPosts
}

// MakeDigestUsers returns text digest about top users for given chat
func MakeDigestUsers(storage Storage, chatID int64) (string, error) {
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
		user := userInfo{id: userID, rating: rating}
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

// MakeDigestPosts returns text digest about top users for given chat
func MakeDigestPosts(storage Storage, chatID int64, chatName string) (string, error) {
	// get users and their rating
	ids, err := storage.Posts.List(chatID)
	if err != nil {
		return "", errorx.Decorate(err, "cannot list posts")
	}
	posts := make([]postInfo, len(ids))
	for _, postID := range ids {
		rating, err := storage.Likes.Count(chatID, postID)
		if err != nil {
			return "", errorx.Decorate(err, "cannot get likes for post")
		}
		post := postInfo{id: postID, rating: rating}
		posts = append(posts, post)
	}

	// sort posts by rating
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].rating > posts[j].rating
	})

	// make top 5 posts
	digest := "*Top 10 posts:*"
	for i, post := range posts[:min(10, len(posts))] {
		if post.rating == 0 {
			continue
		}
		post.title, err = GetPostTitle(chatName, post.id)
		if err != nil {
			return "", errorx.Decorate(err, "cannot get post title", chatName, post.id)
		}
		digest += fmt.Sprintf(
			"\n%d. [%s](https://t.me/%s/%d): %s",
			i+1, post.title, chatName, post.id, ByteCount(post.rating),
		)
	}
	return digest, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetPostTitle extracts post title for public channel
func GetPostTitle(chatName string, postID int) (string, error) {
	if chatName == "" {
		return fmt.Sprintf("#%d", postID), nil
	}
	url := fmt.Sprintf("https://post.tg.dev/%s/%d", chatName, postID)
	// Request the HTML page.
	res, err := http.Get(url)
	if err != nil {
		return "", errorx.Decorate(err, "cannot make request")
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return "", errorx.Decorate(err, "cannot get page content", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", errorx.Decorate(err, "cannot parse HTML")
	}
	text := doc.Find(".tgme_widget_message_text").Text()

	return strings.Split(text, " ")[0], nil
}
