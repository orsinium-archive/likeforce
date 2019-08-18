package storage

import (
	"strconv"

	"github.com/go-redis/redis"
)

// Chat is a main entrypoint for all storage operations
type Chat struct {
	ID     int64
	client *redis.Client
}

// Post returns Post instance for current Chat
func (chat *Chat) Post(id int) Post {
	return Post{ID: id, ChatID: chat.ID, client: chat.client}
}

// Posts returns list of registered posts for Chat
func (chat *Chat) Posts() ([]int, error) {
	key := makeKeyPosts(chat.ID)
	result, err := chat.client.Exists(key).Result()
	if err != nil {
		return nil, err
	}
	if result == 0 {
		return nil, nil
	}
	idsRaw, err := chat.client.SMembers(key).Result()
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
