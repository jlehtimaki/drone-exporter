package repository

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/fatih/structs"

	log "github.com/sirupsen/logrus"

	"github.com/jlehtimaki/drone-exporter/pkg/api"
	influxdb "github.com/jlehtimaki/drone-exporter/pkg/drivers"
)

type Repo struct {
	Id        int
	Name      string
	Slug      string
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
	Created    int64
	Started    int64
	Finished   int64
	Duration   int64
	WaitTime   int64
	Time       time.Time
	RepoName   string
	RepoTeam   string
	Pipeline   string
	BuildState int
}

type BuildInfo struct {
	Id       int
	Sender   string
	Created  int64
	Started  int64
	Finished int64
	Duration int64
	WaitTime int64
	Stages   []struct {
		Os       string
		Arch     string
		Status   string
		Kind     string
		Type     string
		Name     string
		Created  int64
		Started  int64
		Stopped  int64
		Duration int64
		WaitTime int64
		Machine  string
	}
}

// a pipeline with some repo data
type Fields struct {
	RepoSlug    string
	BuildId     int
	BuildSender string
	WaitTime    int64
	Duration    int64
	Os          string
	Arch        string
	Status      string
	Name        string
	Time        time.Time
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
		if err := sendBuildPipelines(repo, build); err != nil {
			return err
		}
	}

	return nil
}

func sendBuildPipelines(repo Repo, build Build) error {

	// Do API Call to Drone
	var subUrlPath = fmt.Sprintf("/api/repos/%s/%s/builds/%s", repo.Namespace, repo.Name, strconv.Itoa(build.Number))
	log.Debugf("[%s] pulling builds: %s", repo.Slug, subUrlPath)
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

	// Loop through build info stages and save the results into DB
	// Don't save running pipelines and set BuildState integer according to the status because of Grafana

	fieldList := []map[string]interface{}{}
	for _, stage := range buildInfo.Stages {
		if stage.Status != "running" {
			fields := &Fields{
				RepoSlug:    repo.Slug,
				BuildId:     build.Id,
				BuildSender: build.Sender,
				WaitTime:    stage.Started - stage.Created,
				Duration:    stage.Stopped - stage.Started,
				Os:          stage.Os,
				Arch:        stage.Arch,
				Status:      stage.Status,
				Name:        stage.Name,
				Time:        time.Unix(stage.Started, 0),
			}

			fieldList = append(fieldList, structs.Map(fields))
		}
	}

	log.Debugf("[%s] sending %d points to influx", repo.Slug, len(fieldList))
	if err := influxdb.RunBatch(fieldList); err != nil {
		return err
	}

	return nil
}
