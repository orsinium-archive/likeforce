package storage

import (
	"github.com/go-redis/redis"
)

// Like is a set of operations to store info about likes on post from user
type Like struct {
	UserID int
	PostID int
	ChatID int64

	client *redis.Client
}

// Create to save a new like for given post
func (like *Like) Create(userID int) (err error) {
	// add post into list of user likes
	err = like.client.SAdd(
		makeKeyUserLikes(like.ChatID, userID),
		makeValuePost(like.ChatID, like.PostID),
	).Err()
	if err != nil {
		return err
	}

	// increment likes count for post
	err = like.client.Incr(makeKeyPostLikes(like.ChatID, like.PostID)).Err()
	if err != nil {
		return err
	}
	return nil
}

// Remove to dislike given post
func (like *Like) Remove(userID int) (err error) {
	err = like.client.SRem(
		makeKeyUserLikes(like.ChatID, userID),
		makeValuePost(like.ChatID, like.PostID),
	).Err()
	if err != nil {
		return err
	}
	err = like.client.Decr(makeKeyPostLikes(like.ChatID, like.PostID)).Err()
	if err != nil {
		return err
	}
	return nil
}

// Exists returns true if post is already liked by user
func (like *Like) Exists(user int) (bool, error) {
	return like.client.SIsMember(
		makeKeyUserLikes(like.ChatID, user),
		makeValuePost(like.ChatID, like.PostID),
	).Result()
}
