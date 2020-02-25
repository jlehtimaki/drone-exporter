package main

import (
    "github.com/jlehtimaki/drone-exporter/pkg/repository"
    "github.com/jlehtimaki/drone-exporter/pkg/env"
    log "github.com/sirupsen/logrus"
    "strconv"
    "time"
)

func main()  {
    // Set logging format
    formatter := &log.TextFormatter{
        FullTimestamp: true,
    }
    log.SetFormatter(formatter)

    // Get loop interval
    interval, err := strconv.Atoi(env.GetEnv("INTERVAL", "2"))
    if err != nil {
        log.Fatal("could not convert INTERVAL value: %s to integer", interval)
    }

    // Start main loop
    for {
        log.Info("Getting data")
        err := repository.GetRepos()
        if err != nil {
            log.Fatalf("error: %s", err)
        }
        log.Infof("Waiting %d minutes", interval)
        time.Sleep(time.Duration(interval) * time.Minute)
    }
}
