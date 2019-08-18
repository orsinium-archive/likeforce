package storage

import (
	"strconv"

	"github.com/go-redis/redis"
)

// User is a Redis operations with user
type User struct {
	ID     int
	ChatID int64

	client *redis.Client
}

// Rating return Rating instance to perform operations with user's rating
func (user *User) Rating() Rating {
	return Rating{user}
}

// Posts to get posts count for user
func (user *User) Posts() ([]Post, error) {
	key := makeKeyUserPosts(user.ChatID, user.ID)

	// check key existence
	keysCount, err := user.client.Exists(key).Result()
	if err != nil {
		return nil, err
	}
	if keysCount == 0 {
		return nil, nil
	}

	// get IDs and convert into objects
	idsRaw, err := user.client.SMembers(key).Result()
	if err != nil {
		return nil, err
	}
	posts := make([]Post, len(idsRaw))
	for _, idRaw := range idsRaw {
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			return nil, err
		}
		posts = append(posts, Post{ID: id, ChatID: user.ChatID, client: user.client})
	}
	return posts, nil
}
