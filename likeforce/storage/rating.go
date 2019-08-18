package storage

// Rating is a Redis operations with user's rating
type Rating struct {
	*User
}

// Incr rating for user
func (user *Rating) Incr() error {
	return user.client.Incr(makeKeyUserRating(user.ChatID, user.ID)).Err()
}

// Decr rating for user
func (user *Rating) Decr() error {
	return user.client.Decr(makeKeyUserRating(user.ChatID, user.ID)).Err()
}

// Get user's rating
func (user *Rating) Get() (int, error) {
	key := makeKeyUserRating(user.ChatID, user.ID)
	keysCount, err := user.client.Exists(key).Result()
	if err != nil {
		return 0, err
	}
	if keysCount == 0 {
		return 0, nil
	}
	return user.client.Get(key).Int()
}
