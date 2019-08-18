package storage

import (
	"github.com/go-redis/redis"
)

// Author is a storage abstraction to process operations on post and post's author
type Author struct {
	UserID int
	PostID int
	ChatID int64

	client *redis.Client
}

// Create to save a new post for given chat
func (author *Author) Create() error {
	// add post into list of posts in the chat
	err := author.client.SAdd(makeKeyPosts(author.ChatID), author.PostID).Err()
	if err != nil {
		return err
	}

	// add post into list of posts of the user
	return author.client.SAdd(
		makeKeyUserPosts(author.ChatID, author.UserID),
		makeValuePost(author.ChatID, author.PostID),
	).Err()
}
