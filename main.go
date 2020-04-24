package main

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	dronecli "github.com/drone/drone-go/drone"
	"github.com/fatih/structs"
	"github.com/jlehtimaki/drone-exporter/pkg/drivers/influxdb"
	"github.com/jlehtimaki/drone-exporter/pkg/drone"
	"github.com/jlehtimaki/drone-exporter/pkg/env"
	log "github.com/sirupsen/logrus"
)

var logLevel = env.GetEnv("LOG_LEVEL", "error")
var driver = env.GetEnv("DRIVER", "influxdb")

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

	// generate driver client from env vars for error checking
	switch driver {
	case "influxdb":
		_, err := influxdb.GetClient()
		if err != nil {
			log.Fatal(err)
		}

		defer func() {
			err := influxdb.Close()
			if err != nil {
				log.Error(err)
			}
		}()
	}

	droneClient := drone.GetClient()
	if droneClient == nil {
		log.Fatal(errors.New("unable to create drone client, please check env DRONE_URL and DRONE_TOKEN"))
	}

	cli := *droneClient

	// Get loop interval
	interval, err := strconv.Atoi(env.GetEnv("INTERVAL", "2"))
	if err != nil {
		log.Fatal("could not convert INTERVAL value: %s to integer", interval)
	}

	// Start main loop

	for {
		log.Info("Getting Repos")
		repos, err := cli.RepoList()
		if err != nil {
			log.Fatal(err)
		}

		var buildFields []map[string]interface{}
		for _, repo := range repos {
			builds, err := cli.BuildList(repo.Namespace, repo.Name, dronecli.ListOptions{})
			if err != nil {
				log.Fatal(err)
			}

			if len(builds) == 0 {
				continue
			}

			log.Debugf("[%s] processing %d builds", repo.Slug, len(builds))
			var stageFields []map[string]interface{}
			var stepFields []map[string]interface{}
			for _, build := range builds {
				buildInfo, err := cli.Build(repo.Namespace, repo.Name, int(build.Number))
				if err != nil {
					log.Fatal(err)
				}

				buildFields = append(buildFields, structs.Map(&influxdb.Build{
					Time:     time.Unix(buildInfo.Started, 0),
					Number:   buildInfo.Number,
					WaitTime: buildInfo.Started - buildInfo.Created,
					Duration: buildInfo.Finished - buildInfo.Started,
					Source:   buildInfo.Source,
					Target:   buildInfo.Target,
					Started:  buildInfo.Started,
					Created:  buildInfo.Created,
					Finished: buildInfo.Finished,
					Tags: map[string]string{
						"Slug":    repo.Slug,
						"BuildId": fmt.Sprintf("build-%d", buildInfo.Number),
					},
				}))

				for _, stage := range buildInfo.Stages {
					// Loop through build info stages and save the results into DB
					// Don't save running pipelines and set BuildState integer according to the status because of Grafana
					if stage.Status != "running" {
						stageFields = append(stageFields, structs.Map(&influxdb.Stage{
							Time:     time.Unix(stage.Started, 0),
							WaitTime: stage.Started - stage.Created,
							Duration: stage.Stopped - stage.Started,
							OS:       stage.OS,
							Arch:     stage.Arch,
							Status:   stage.Status,
							Name:     stage.Name,
							Tags: map[string]string{
								"Slug":    repo.Slug,
								"BuildId": fmt.Sprintf("build-%d", build.Number),
								"Sender":  build.Sender,
							},
						}))
					}

					for _, step := range stage.Steps {
						stepFields = append(stepFields, structs.Map(&influxdb.Step{
							Time:     time.Unix(step.Started, 0),
							Duration: step.Stopped - step.Started,
							Name:     step.Name,
							Status:   step.Status,
							Tags: map[string]string{
								"Slug":    repo.Slug,
								"BuildId": fmt.Sprintf("build-%d", build.Number),
								"Sender":  build.Sender,
							},
						}))
					}
				}
			}

			if len(stageFields) > 0 {
				log.Debugf("[%s] sending %d stages to db", repo.Slug, len(stageFields))
				go func() {
					err = influxdb.Batch("stages", stageFields)
					if err != nil {
						log.Error(err)
					}
				}()
			}

			if len(stepFields) > 0 {
				log.Debugf("[%s] sending %d steps to db", repo.Slug, len(stepFields))
				go func() {
					err = influxdb.Batch("steps", stepFields)
					if err != nil {
						log.Error(err)
					}
				}()
			}
		}

		if len(buildFields) > 0 {
			log.Debugf("sending %d builds to db", len(buildFields))
			go func() {
				err = influxdb.Batch("builds", buildFields)
				if err != nil {
					log.Error(err)
				}
			}()
		}

		log.Infof("Waiting %d minutes", interval)
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}
