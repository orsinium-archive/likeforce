package main

import (
	"os"

	"github.com/francoispqt/onelog"
	"github.com/orsinium/likeforce/likeforce"
	"github.com/orsinium/likeforce/likeforce/storage"
)

func main() {
	logger := onelog.New(os.Stdout, onelog.ALL)
	config, err := likeforce.ReadConfig("config.toml")
	if err != nil {
		logger.FatalWith("cannot read config").Err("error", err).Write()
		return
	}
	storage, err := storage.NewStorage(config.Redis)
	if err != nil {
		logger.FatalWith("cannot create Redis connection").Err("error", err).Write()
		return
	}
	tg, err := likeforce.NewTelegram(config, storage, logger)
	if err != nil {
		logger.FatalWith("cannot create Telegram connection").Err("error", err).Write()
		return
	}
	err = tg.Serve()
	if err != nil {
		logger.FatalWith("cannot connect to telegram").Err("error", err).Write()
		return
	}
}
