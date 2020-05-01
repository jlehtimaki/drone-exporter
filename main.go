package main

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	dronecli "github.com/drone/drone-go/drone"
	"github.com/jlehtimaki/drone-exporter/pkg/drivers/influxdb"
	"github.com/jlehtimaki/drone-exporter/pkg/drone"
	"github.com/jlehtimaki/drone-exporter/pkg/env"
	"github.com/jlehtimaki/drone-exporter/pkg/types"
	log "github.com/sirupsen/logrus"
)

const pageSize = 25

var logLevel = env.GetEnv("LOG_LEVEL", "error")
var envInterval = env.GetEnv("INTERVAL", "2")
var envThreads = env.GetEnv("THREADS", "10")
var cli dronecli.Client

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

	// initialize the influx client
	driver, err := influxdb.NewDriver()
	if err != nil {
		log.Fatal(err)
	}
	defer driver.Close()

	droneClient := drone.GetClient()
	if droneClient == nil {
		log.Fatal(errors.New("unable to create drone client, please check env DRONE_URL and DRONE_TOKEN"))
	}
	cli = *droneClient

	// Get loop interval
	interval, err := strconv.Atoi(envInterval)
	if err != nil {
		log.Fatal("could not convert INTERVAL value: %s to integer", envInterval)
	}

	threads, err := strconv.Atoi(envThreads)
	if err != nil {
		log.Fatal("could not convert THREADS value: %s to integer", envThreads)
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, threads)

	// Start main loop
	for {

		repos, err := cli.RepoList()
		if err != nil {
			log.Fatal(err)
		}

		log.Infof("[drone-exporter] processing %d repos", len(repos))
		wg.Add(len(repos))
		for _, repo := range repos {
			r := repo
			go func() {
				log.Debugf("[%s] starting thread", r.Slug)
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()
				points := processRepo(r, driver.LastBuildNumber(r.Slug))
				if len(points) > 0 {
					log.Debugf("[%s] sending %d points to db", r.Slug, len(points))
					err = driver.Batch(points)
					if err != nil {
						log.Error(err)
					}
				}
				log.Debugf("[%s] thread complete", r.Slug)
			}()
		}
		wg.Wait()

		log.Infof("[drone-exporter] waiting %d minutes", interval)
		time.Sleep(time.Duration(interval) * time.Minute)
	}
}

func processRepo(repo *dronecli.Repo, lastBuildId int64) []types.Point {
	var points []types.Point

	// process first page
	page := 1
	builds, err := cli.BuildList(repo.Namespace, repo.Name, dronecli.ListOptions{
		Page: page,
		Size: pageSize,
	})
	if err != nil {
		log.Fatal(err)
	}

	if len(builds) == 0 {
		log.Debugf("[%s] found zero builds, skipping...", repo.Name)
		return []types.Point{}
	}

	if builds[0].Number == lastBuildId {
		log.Debugf("[%s] found no new builds, skipping...", repo.Name)
		return []types.Point{}
	}

	points = append(points, processBuilds(repo, builds)...)
	if len(builds) < pageSize {
		return []types.Point{} //no pages
	}

	// paginate
	for len(builds) > 0 {
		page++
		builds, err = cli.BuildList(repo.Namespace, repo.Name, dronecli.ListOptions{
			Page: page,
			Size: pageSize,
		})
		if err != nil {
			log.Fatal(err)
		}

		if len(builds) == 0 {
			continue
		}
		ps := processBuilds(repo, builds)
		points = append(points, ps...)
		if len(builds) < pageSize {
			continue
		}
	}

	return points
}

func processBuilds(repo *dronecli.Repo, builds []*dronecli.Build) []types.Point {
	log.Debugf("[%s] processing %d builds", repo.Slug, len(builds))
	var points []types.Point
	for _, build := range builds {
		buildInfo, err := cli.Build(repo.Namespace, repo.Name, int(build.Number))
		if err != nil {
			log.Fatal(err)
		}

		if buildInfo.Status == "running" {
			continue
		}

		var waittime int64
		if buildInfo.Started == 0 {
			waittime = buildInfo.Updated - buildInfo.Created
		} else {
			waittime = buildInfo.Started - buildInfo.Created
		}

		var duration int64
		if buildInfo.Finished == 0 {
			duration = buildInfo.Updated - buildInfo.Started
		} else {
			duration = buildInfo.Finished - buildInfo.Started
		}

		points = append(points, &types.Build{
			Time:     time.Unix(buildInfo.Started, 0),
			Number:   buildInfo.Number,
			Status:   buildInfo.Status,
			WaitTime: waittime,
			Duration: duration,
			Source:   buildInfo.Source,
			Target:   buildInfo.Target,
			Started:  buildInfo.Started,
			Created:  buildInfo.Created,
			Finished: buildInfo.Finished,
			BuildId:  build.Number,
			Tags: map[string]string{
				"DroneAddress": drone.GetHost(),
				"Slug":         repo.Slug,
				"Status":       buildInfo.Status,
				"BuildId":      fmt.Sprintf("build-%d", buildInfo.Number),
			},
		})

		for _, stage := range buildInfo.Stages {
			// Loop through build info stages and save the results into DB
			// Don't save running pipelines and set BuildState integer according to the status because of Grafana
			var waittime int64
			if stage.Started == 0 {
				waittime = stage.Updated - stage.Created
			} else {
				duration = stage.Started - stage.Created
			}

			var duration int64
			if stage.Stopped == 0 {
				duration = stage.Updated - stage.Started
			} else {
				duration = stage.Stopped - stage.Started
			}

			points = append(points, &types.Stage{
				Time:     time.Unix(stage.Started, 0),
				WaitTime: waittime,
				Duration: duration,
				OS:       stage.OS,
				Arch:     stage.Arch,
				Status:   stage.Status,
				Name:     stage.Name,
				BuildId:  build.Number,
				Tags: map[string]string{
					"DroneAddress": drone.GetHost(),
					"Slug":         repo.Slug,
					"BuildId":      fmt.Sprintf("build-%d", build.Number),
					"Sender":       build.Sender,
					"Name":         stage.Name,
					"OS":           stage.OS,
					"Arch":         stage.Arch,
					"Status":       stage.Status,
				},
			})

			for _, step := range stage.Steps {
				duration := step.Stopped - step.Started
				if duration < 0 {
					duration = 0
				}
				points = append(points, &types.Step{
					Time:     time.Unix(step.Started, 0),
					Duration: duration,
					Name:     step.Name,
					Status:   step.Status,
					Tags: map[string]string{
						"DroneAddress": drone.GetHost(),
						"Slug":         repo.Slug,
						"BuildId":      fmt.Sprintf("build-%d", build.Number),
						"Sender":       build.Sender,
						"Name":         step.Name,
						"Status":       step.Status,
					},
				})
			}
		}
	}

	return points
}
