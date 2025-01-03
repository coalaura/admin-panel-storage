package main

import "github.com/coalaura/logger"

var (
	log = logger.New()

	config  *Config
	storage *Storage
)

func main() {
	var err error

	log.Info("Reading config...")
	config, err = ReadConfig()
	log.MustPanic(err)

	log.Info("Initializing storage...")
	storage, err = NewStorage(config.Root)
	log.MustPanic(err)

	log.Info("Starting TCP server...")
	log.MustPanic(StartTCPServer())
}
