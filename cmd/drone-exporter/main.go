package main

import (
	"os"
	"strconv"
	"time"

	"github.com/jlehtimaki/drone-exporter/pkg/env"
	"github.com/jlehtimaki/drone-exporter/pkg/repository"
	log "github.com/sirupsen/logrus"
)

var logLevel = env.GetEnv("LOG_LEVEL", "error")

func main() {
	// Set logging format
	formatter := &log.TextFormatter{
		FullTimestamp: true,
	}
	log.SetFormatter(formatter)

	switch logLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	}

	// Get loop interval
	interval, err := strconv.Atoi(env.GetEnv("INTERVAL", "2"))
	if err != nil {
		log.Fatal("could not convert INTERVAL value: %s to integer", interval)
	}

	// Start main loop
	for {
		log.Info("Getting data")
		if err := repository.GetRepos(); err != nil {
			log.Errorf("error: %s", err)
			os.Exit(1)
		}
		log.Infof("Waiting %d minutes", interval)
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
