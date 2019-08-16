package likeforce

import (
	"fmt"

	"github.com/go-redis/redis"
)

// Users is a set of operations to store user stat in Redis
type Users struct {
	client *redis.Client
}

func makeKeyPosts(chat int64, user int) string {
	return fmt.Sprintf("likes:user:posts:%d:%d", chat, user)
}

func makeKeyRating(chat int64, user int) string {
	return fmt.Sprintf("likes:user:rating:%d:%d", chat, user)
}

// AddPost to increment posts count for user
func (storage *Users) AddPost(chat int64, user int) error {
	return storage.client.Incr(makeKeyPosts(chat, user)).Err()
}

// AddRating to increment rating for user
func (storage *Users) AddRating(chat int64, user int) error {
	return storage.client.Incr(makeKeyRating(chat, user)).Err()
}

// RemoveRating to decrement rating for user
func (storage *Users) RemoveRating(chat int64, user int) error {
	return storage.client.Decr(makeKeyRating(chat, user)).Err()
}

// PostsCount to get posts count for user
func (storage *Users) PostsCount(chat int64, user int) (int, error) {
	return storage.client.Get(makeKeyPosts(chat, user)).Int()
}

// RatingCount to get rating for user
func (storage *Users) RatingCount(chat int64, user int) (int, error) {
	return storage.client.Get(makeKeyRating(chat, user)).Int()
}

// ByteCount to make human-readable rating
func ByteCount(count int) string {
	const unit = 1000
	if count < unit {
		return fmt.Sprintf("%d", count)
	}
	div, exp := int64(unit), 0
	for n := count / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%c", float64(count)/float64(div), "kMGTPE"[exp])
}

// Stat to get human-readable message with user stat
func (storage *Users) Stat(chat int64, user int) (string, error) {
	posts, err := storage.PostsCount(chat, user)
	if err != nil {
		return "", err
	}
	rating, err := storage.RatingCount(chat, user)
	if err != nil {
		return "", err
	}
	if posts == 0 {
		return "First blood!", nil
	}
	const tmpl = "user stat:\nposts: %s\nrating: %s"
	return fmt.Sprintf(tmpl, ByteCount(posts), ByteCount(rating)), nil
}

// NewUsers creates Users with a new Redis connection
func NewUsers(config redis.Options) (Users, error) {
	client := redis.NewClient(&config)
	_, err := client.Ping().Result()
	if err != nil {
		return Users{}, err
	}
	return Users{client: client}, nil
}
