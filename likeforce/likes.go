package likeforce

import (
	"fmt"

	"github.com/go-redis/redis"
)

// Likes is a set of operations to store info about likes in Redis
type Likes struct {
	client *redis.Client
}

func makeKeyPost(chat int64, post int) string {
	return fmt.Sprintf("likes:post:%d:%d", chat, post)
}

func makeKeyUser(chat int64, user int) string {
	return fmt.Sprintf("likes:user:%d:%d", chat, user)
}

func makeValuePost(chat int64, post int) string {
	return fmt.Sprintf("%d:%d", chat, post)
}

// Add to save a new like for given post
func (storage *Likes) Add(chat int64, post, user int) (err error) {
	err = storage.client.SAdd(
		makeKeyUser(chat, user),
		makeValuePost(chat, post),
	).Err()
	if err != nil {
		return err
	}
	err = storage.client.Incr(makeKeyPost(chat, post)).Err()
	if err != nil {
		return err
	}
	return nil
}

// Remove to dislike given post
func (storage *Likes) Remove(chat int64, post, user int) (err error) {
	err = storage.client.SRem(
		makeKeyUser(chat, user),
		makeValuePost(chat, post),
	).Err()
	if err != nil {
		return err
	}
	err = storage.client.Decr(makeKeyPost(chat, post)).Err()
	if err != nil {
		return err
	}
	return nil
}

// Has returns true if post is already liked by user
func (storage *Likes) Has(chat int64, post, user int) (bool, error) {
	return storage.client.SIsMember(
		makeKeyUser(chat, user),
		makeValuePost(chat, post),
	).Result()
}

// Count returns count of likes for a given post
func (storage *Likes) Count(chat int64, post int) (int, error) {
	key := makeKeyPost(chat, post)
	result, err := storage.client.Exists(key).Result()
	if err != nil {
		return 0, err
	}
	if result == 0 {
		return 0, nil
	}
	return storage.client.Get(key).Int()
}
