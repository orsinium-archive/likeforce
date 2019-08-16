package likeforce

import (
	"github.com/BurntSushi/toml"
	"github.com/go-redis/redis"
)

type TelegramConfig struct {
	Token   string `toml:"token"`
	Debug   bool   `toml:"debug"`
	Timeout int    `toml:"timeout"`
}

// type RedisConfig struct {
// 	Addr     string `toml:"address"`
// 	Password string `toml:"password"`
// 	DB       int    `toml:"database"`
// }

type Config struct {
	Telegram TelegramConfig
	Redis    redis.Options
}

func ReadConfig(path string) (Config, error) {
	var config Config
	_, err := toml.DecodeFile(path, &config)
	return config, err
}
