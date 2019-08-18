package storage

import (
	"strconv"
	"strings"

	"github.com/go-redis/redis"
)

// Chat is a main entrypoint for all storage operations
type Chat struct {
	ID     int64
	client *redis.Client
}

// Post returns Post instance for current Chat
func (chat *Chat) Post(id int) *Post {
	return &Post{ID: id, ChatID: chat.ID, client: chat.client}
}

// User returns User instance for current Chat
func (chat *Chat) User(id int) *User {
	return &User{ID: id, ChatID: chat.ID, client: chat.client}
}

// Posts returns list of registered posts for Chat
func (chat *Chat) Posts() ([]Post, error) {
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
	posts := make([]Post, len(idsRaw))
	for i, idRaw := range idsRaw {
		id, err := strconv.Atoi(idRaw)
		if err != nil {
			return nil, err
		}
		posts[i] = Post{ID: id, ChatID: chat.ID, client: chat.client}
	}
	return posts, nil
}

// Users returns list of users with posts in the Chat
func (chat *Chat) Users() (users []User, err error) {
	// get keys
	pattern := strings.Replace(makeKeyUserPosts(chat.ID, 0), "user-0", "user-*", 1)
	keys, err := chat.client.Keys(pattern).Result()
	if err != nil {
		return
	}

	// extract users IDs from keys
	var userID int
	for _, key := range keys {
		part := strings.Split(strings.Split(key, "user-")[1], ":")[0]
		userID, err = strconv.Atoi(part)
		if err != nil {
			return
		}
		users = append(users, User{ID: userID, ChatID: chat.ID, client: chat.client})
	}
	return
}
