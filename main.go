package main

import (
	"os"

	"github.com/francoispqt/onelog"
	"github.com/orsinium/likeforce/likeforce"
)

func main() {
	logger := onelog.New(os.Stdout, onelog.ALL)
	config, err := likeforce.ReadConfig("config.toml")
	if err != nil {
		logger.FatalWith("cannot read config").Err("error", err).Write()
		return
	}
	likes, err := likeforce.NewLikes(config.Redis)
	if err != nil {
		logger.FatalWith("cannot create Redis connection").Err("error", err).Write()
		return
	}
	tg, err := likeforce.NewTelegram(config.Telegram, likes)
	if err != nil {
		logger.FatalWith("cannot create Telegram connection").Err("error", err).Write()
		return
	}
	logger.Info("Serve")
	tg.Serve()
}
