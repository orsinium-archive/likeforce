package storage

import "github.com/go-redis/redis"

// Storage is a collection of Redis data managers
type Storage struct {
	client *redis.Client
}

// Chat creates a new Chat instance with current Redis connection
func (storage *Storage) Chat(chatID int64) *Chat {
	return &Chat{ID: chatID, client: storage.client}
}

// NewStorage creates a Storage instance with a new Redis connection
func NewStorage(config redis.Options) (Storage, error) {
	client := redis.NewClient(&config)
	_, err := client.Ping().Result()
	if err != nil {
		return Storage{}, err
	}
	storage := Storage{client}
	return storage, nil
}
