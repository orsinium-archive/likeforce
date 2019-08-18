package likeforce

import (
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
)

// TelegramConfig is a container for telegram configuration
type TelegramConfig struct {
	Token   string
	Debug   bool
	Timeout int
	Admin   string
}

// MessagesConfig is a container for bot messages customization
type MessagesConfig struct {
	Like     []string
	Liked    string
	Disliked string
	Error    string
}

// Config is a container for TOML config content
type Config struct {
	Telegram TelegramConfig
	Messages MessagesConfig
	Redis    redis.Options
}

// ReadConfig reads TOML config into Config
func ReadConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return config, err
}
