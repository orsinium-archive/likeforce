package likeforce

import (
	"fmt"
	"strconv"
	"strings"

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

func makeKeyName(user int) string {
	return fmt.Sprintf("likes:user:login:%d", user)
}

// AddName to save username
func (storage *Users) AddName(user int, name string) error {
	return storage.client.Set(makeKeyName(user), name, 0).Err()
}

// GetName to get username
func (storage *Users) GetName(user int) (string, error) {
	return storage.client.Get(makeKeyName(user)).Result()
}

// List returns list of registered users IDs
func (storage *Users) List(chat int64) (users []int, err error) {
	// get keys
	pattern := fmt.Sprintf("likes:user:rating:%d:*", chat)
	keys, err := storage.client.Keys(pattern).Result()
	if err != nil {
		return
	}

	// extract users IDs from keys
	var userID int
	for _, key := range keys {
		parts := strings.Split(key, ":")
		userID, err = strconv.Atoi(parts[len(parts)-1])
		if err != nil {
			return
		}
		users = append(users, userID)
	}
	return
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
