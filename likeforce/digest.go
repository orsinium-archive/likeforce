package likeforce

import (
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/joomcode/errorx"
	"github.com/orsinium/likeforce/likeforce/storage"
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

// MakeDigestUsers returns text digest about top users for given chat
func MakeDigestUsers(chat *storage.Chat) (string, error) {
	// get users and their rating
	rawUsers, err := chat.Users()
	if err != nil {
		return "", errorx.Decorate(err, "cannot get users")
	}
	users := make([]userInfo, len(rawUsers))
	for i, rawUser := range rawUsers {
		rating, err := rawUser.Rating().Get()
		if err != nil {
			return "", errorx.Decorate(err, "cannot get user rating (user %d)", rawUser.ID)
		}
		users[i] = userInfo{id: rawUser.ID, rating: rating}
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
		user.name, err = chat.User(user.id).Name()
		if err != nil {
			return "", errorx.Decorate(err, "cannot get user name (user %d)", user.id)
		}
		digest += fmt.Sprintf("\n%d. [%s](tg://user?id=%d): (%s)", i+1, user.name, user.id, ByteCount(user.rating))
	}
	return digest, nil
}

// MakeDigestPosts returns text digest about top users for given chat
func MakeDigestPosts(chat *storage.Chat, chatName string) (string, error) {
	// get users and their rating
	rawPosts, err := chat.Posts()
	if err != nil {
		return "", errorx.Decorate(err, "cannot list posts")
	}
	posts := make([]postInfo, len(rawPosts))
	for i, rawPost := range rawPosts {
		rating, err := rawPost.Likes()
		if err != nil {
			return "", errorx.Decorate(err, "cannot get likes for post")
		}
		posts[i] = postInfo{id: rawPost.ID, rating: rating}
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
			return "", errorx.Decorate(err, "cannot get post title (chat %s, post %d)", chatName, post.id)
		}
		digest += fmt.Sprintf(
			"\n%d. [%s](https://t.me/%s/%d) (%s)",
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
		return "", errorx.Decorate(err, "cannot get page content (%d %s)", res.StatusCode, res.Status)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return "", errorx.Decorate(err, "cannot parse HTML")
	}
	article := doc.Find(".js-message_text")
	html, err := article.Html()
	if err != nil {
		return "", errorx.Decorate(err, "cannot extract HTML element")
	}
	text := article.Text()

	// extract project name from link preview title
	title := doc.Find(".link_preview_title").First().Text()
	if title != "" {
		return title, nil
	}

	// extract project name from github URL
	rexGitHub, err := regexp.Compile(`github\.com/[a-zA-Z\d\-]+/([a-zA-Z\d\.\-\_]+)`)
	if err != nil {
		return "", errorx.Decorate(err, "cannot compile regexp")
	}
	if rexGitHub.MatchString(html) {
		projectName := rexGitHub.FindStringSubmatch(html)[1]
		return projectName, nil
	}

	// extract project name from the first link
	projectName := article.Find("a").First().Text()
	if projectName != "" {
		return projectName, nil
	}

	// just return the first word
	return strings.Split(text, " ")[0] + "...", nil
}
