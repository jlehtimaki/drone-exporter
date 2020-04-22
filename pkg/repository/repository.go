package repository

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fatih/structs"
	"github.com/jlehtimaki/drone-exporter/pkg/api"
	influxdb "github.com/jlehtimaki/drone-exporter/pkg/drivers"
)

type Repo struct {
	Id        int
	Name      string
	Active    bool
	Namespace string
}

type Build struct {
	Id         int
	Trigger    string
	Status     string
	Number     int
	Event      string
	Action     string
	Link       string
	Message    string
	Ref        string
	Source     string
	Target     string
	Sender     string
	Started    int64
	Finished   int64
	Time       time.Time
	RepoName   string
	RepoTeam   string
	Pipeline   string
	BuildState int
}

type BuildInfo struct {
	Id       int
	Started  int64
	Finished int64
	Duration int64
	Stages   []struct {
		Status   string
		Kind     string
		Type     string
		Name     string
		Started  int64
		Stopped  int64
		Duration int64
	}
}

func GetRepos() error {
	defer func() {
		err := influxdb.Close()
		if err != nil {
			log.Error(err)
		}
	}()

	var subUrlPath = "/api/user/repos"
	data, err := api.ApiRequest(subUrlPath)
	if err != nil {
		return fmt.Errorf("error getting repositories: %w", err)
	}

	var repos []Repo
	if err := json.Unmarshal([]byte(data), &repos); err != nil {
		return fmt.Errorf("could not create repos struct: %w", err)
	}

	log.Debugf("processing %d repos", len(repos))
	for _, v := range repos {
		if v.Active {
			log.Debugf("[%s] finding builds", v.Name)
			if err := getBuilds(v); err != nil {
				return err
			}
		}
	}
	return nil
}

func getBuilds(repo Repo) error {
	// Setup API url path
	subUrlPath := fmt.Sprintf("/api/repos/%s/%s/builds", repo.Namespace, repo.Name)

	// Do the API call to get Build data
	data, err := api.ApiRequest(subUrlPath)
	if err != nil {
		return fmt.Errorf("could not get builds: %w", err)
	}

	// Get builds for the repo
	var builds []Build
	if err := json.Unmarshal([]byte(data), &builds); err != nil {
		return fmt.Errorf("could not create build struct: %w", err)
	}

	log.Debugf("[%s] found %d builds", repo.Name, len(builds))
	// Loop through Builds and get more detailed information
	for _, build := range builds {
		if err := getBuildInfo(build, repo.Namespace, repo.Name); err != nil {
			return err
		}
	}

	return nil
}

func getBuildInfo(build Build, repoNamespace string, repoName string) error {
	// Set build variables
	build.Time = time.Unix(build.Started, 0)
	build.RepoTeam = repoNamespace
	build.RepoName = repoName

	// Do API Call to Drone
	var subUrlPath = fmt.Sprintf("/api/repos/%s/%s/builds/%s", repoNamespace, repoName, strconv.Itoa(build.Number))
	log.Debugf("[%s] pulling builds: %s", repoName, subUrlPath)
	data, err := api.ApiRequest(subUrlPath)
	if err != nil {
		return fmt.Errorf("error getting buildinfo: %w", err)
	}

	// Create struct of API Request data
	// Set empty BuildInfo Struct
	var buildInfo BuildInfo
	if err := json.Unmarshal([]byte(data), &buildInfo); err != nil {
		return fmt.Errorf("could not create buildinfo struct: %w", err)
	}

	buildInfo.Duration = buildInfo.Finished - buildInfo.Started
	log.Debugf("[%s] build %d took %d", repoName, buildInfo.Id, buildInfo.Duration)
	// Loop through build info stages and save the results into DB
	// Don't save running pipelines and set BuildState integer according to the status because of Grafana
	for _, y := range buildInfo.Stages {
		if y.Status != "running" {
			if y.Status == "success" {
				build.BuildState = 1
				build.Status = "success"
			} else {
				build.BuildState = 0
				build.Status = "failure"
			}
			y.Duration = y.Stopped - y.Started
			build.Pipeline = y.Name

			log.Debugf("[%s] sending %s to influx: time %d duration %d", repoName, y.Name, build.Time.Unix(), y.Duration)
			if err := influxdb.Run(structs.Map(build), y.Name); err != nil {
				return err
			}
		}
	}

	return nil
}
