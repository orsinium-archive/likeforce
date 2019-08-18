package storage

import (
	"github.com/go-redis/redis"
)

// Post is a set of operations to store info about posts in Redis
type Post struct {
	ID     int
	ChatID int64

	client *redis.Client
}

// Like returns Like instance for current Post and given User
func (post *Post) Like(userID int) *Like {
	return &Like{PostID: post.ID, ChatID: post.ChatID, UserID: userID, client: post.client}
}

// Author returns Author instance for current Post and given User
func (post *Post) Author(userID int) *Author {
	return &Author{PostID: post.ID, ChatID: post.ChatID, UserID: userID, client: post.client}
}

// AuthorID returns Author's ID for current Post
func (post *Post) AuthorID() (int, error) {
	key := makeKeyPostAuthor(post.ChatID, post.ID)
	return post.client.Get(key).Int()
}

// Exists returns true if post is already registered
func (post *Post) Exists() (bool, error) {
	key := makeKeyPosts(post.ChatID)
	result, err := post.client.Exists(key).Result()
	if err != nil {
		return false, err
	}
	if result == 0 {
		return false, nil
	}
	return post.client.SIsMember(key, post.ID).Result()
}

// Likes returns count of likes for a post
func (post *Post) Likes() (int, error) {
	key := makeKeyPostLikes(post.ChatID, post.ID)
	result, err := post.client.Exists(key).Result()
	if err != nil {
		return 0, err
	}
	if result == 0 {
		return 0, nil
	}
	return post.client.Get(key).Int()
}
