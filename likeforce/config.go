package likeforce

import (
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
)

type TelegramConfig struct {
	Token   string
	Debug   bool
	Timeout int
}

type MessagesConfig struct {
	Like     string
	Liked    string
	Disliked string
	Error    string
}

// type RedisConfig struct {
// 	Addr     string `toml:"address"`
// 	Password string `toml:"password"`
// 	DB       int    `toml:"database"`
// }

type Config struct {
	Telegram TelegramConfig
	Messages MessagesConfig
	Redis    redis.Options
}

func ReadConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return config, err
}
