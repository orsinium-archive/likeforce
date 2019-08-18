package likeforce

import "github.com/go-redis/redis"

// Storage is a collection of Redis data managers
type Storage struct {
	Likes Likes
	Posts Posts
	Users Users
}

// NewStorage creates a Storage instance with a new Redis connection
func NewStorage(config redis.Options) (Storage, error) {
	client := redis.NewClient(&config)
	_, err := client.Ping().Result()
	if err != nil {
		return Storage{}, err
	}
	storage := Storage{
		Likes: Likes{client},
		Posts: Posts{client},
		Users: Users{client},
	}
	return storage, nil
}
