package likeforce

import (
	"fmt"
	"strconv"

	"github.com/go-redis/redis"
)

// Posts is a set of operations to store info about posts in Redis
type Posts struct {
	client *redis.Client
}

func makeKeyChat(chat int64) string {
	return fmt.Sprintf("likes:posts:%d", chat)
}

// Add to save a new post for given chat
func (storage *Posts) Add(chat int64, post int) (err error) {
	return storage.client.SAdd(
		makeKeyChat(chat),
		post,
	).Err()
}

// Has returns true if post is already added in channel
func (storage *Posts) Has(chat int64, post int) (bool, error) {
	key := makeKeyChat(chat)
	result, err := storage.client.Exists(key).Result()
	if err != nil {
		return false, err
	}
	if result == 0 {
		return false, nil
	}
	return storage.client.SIsMember(key, post).Result()
}

// List returns list of registered posts IDs
func (storage *Posts) List(chat int64) ([]int, error) {
	key := makeKeyChat(chat)
	result, err := storage.client.Exists(key).Result()
	if err != nil {
		return nil, err
	}
	if result == 0 {
		return nil, nil
	}
	idsRaw, err := storage.client.SMembers(key).Result()
	if err != nil {
		return nil, err
	}
	ids := make([]int, len(idsRaw))
	for _, idRaw := range idsRaw {
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
