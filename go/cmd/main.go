package main

import (
	"fmt"
	"os"

	"github.com/cgund98/go-eventsrc-example/internal/infra/config"
	"github.com/cgund98/go-eventsrc-example/internal/infra/logging"
)

func main() {

	config, err := config.LoadConfig()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	logging.Logger.Info("Starting order service...", "config", config)
}
