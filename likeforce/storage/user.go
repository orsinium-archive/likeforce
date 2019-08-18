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
func (user *User) Rating() *Rating {
	return &Rating{user}
}

// SetName to save username
func (user *User) SetName(name string) error {
	return user.client.Set(makeKeyUserName(user.ID), name, 0).Err()
}

// Name to get username
func (user *User) Name() (string, error) {
	return user.client.Get(makeKeyUserName(user.ID)).Result()
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
	for i, idRaw := range idsRaw {
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			return nil, err
		}
		posts[i] = Post{ID: id, ChatID: user.ChatID, client: user.client}
	}
	return posts, nil
}
