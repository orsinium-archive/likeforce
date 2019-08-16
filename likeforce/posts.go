package likeforce

import (
	"fmt"

	"github.com/go-redis/redis"
)

// Posts is a set of operations to store info about posts in Redis
type Posts struct {
	client *redis.Client
}

func makeKeyChat(chat int) string {
	return fmt.Sprintf("likes:posts:%d", chat)
}

// Add to save a new post for given chat
func (storage *Posts) Add(chat, post int) (err error) {
	return storage.client.SAdd(
		makeKeyChat(chat),
		post,
	).Err()
}

// Has returns true if post is already added in channel
func (storage *Posts) Has(chat, post int) (bool, error) {
	return storage.client.SIsMember(
		makeKeyChat(chat),
		post,
	).Result()
}

// NewPosts creates Posts with a new Redis connection
func NewPosts(config redis.Options) (Posts, error) {
	client := redis.NewClient(&config)
	_, err := client.Ping().Result()
	if err != nil {
		return Posts{}, err
	}
	return Posts{client: client}, nil
}
