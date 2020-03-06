package repository

import (
  "encoding/json"
  "fmt"
  "github.com/fatih/structs"
  "github.com/jlehtimaki/drone-exporter/pkg/api"
  influxdb "github.com/jlehtimaki/drone-exporter/pkg/drivers"
  "strconv"
  "time"
)

type Repo struct {
  Id          int     `json:"Id"`
  Name        string  `json:"Name"`
  Active      bool	  `json:"Active"`
  Namespace   string  `json:"Namespace"`
}

type Build struct {
  Id          int       `json:"Id"`
  Trigger     string    `json:"Trigger"`
  Status      string    `json:"Status"`
  Number      int       `json:"Number"`
  Event       string    `json:"Event"`
  Action      string    `json:"Action"`
  Link        string    `json:"Link"`
  Message     string    `json:"Message"`
  Ref         string    `json:"Ref"`
  Source      string    `json:"Source"`
  Target      string    `json:"Target"`
  Sender      string    `json:"Sender"`
  Started     int64     `json:"Started"`
  Finished    int64     `json:"Finished"`
  Time        time.Time
  RepoName    string
  RepoTeam    string
  Pipeline    string
  BuildState  int
}

type BuildInfo struct {
  Stages        []struct  {
    Status      string    `json:"Status"`
    Kind        string    `json:"Kind"`
    Type        string    `json:"Type"`
    Name        string    `json:"Name"`
    Started     int       `json:"Started"`
    Finished    int       `json:"Finished"`
  }
}


func GetRepos() error {
  var subUrlPath = "/api/user/repos"
  data, err := api.ApiRequest(subUrlPath)
  if err != nil {
    return fmt.Errorf("error getting repositories: %w", err)
  }

  var repos []Repo
  if err := json.Unmarshal([]byte(data), &repos); err != nil {
    return fmt.Errorf( "could not create repos struct: %w", err)
  }
  for _, v := range repos {
    if v.Active {
      err := getBuilds(v)
      if err != nil {
        return fmt.Errorf("could not get builds: %w", err)
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
  if err:= json.Unmarshal([]byte(data), &builds); err != nil {
    return fmt.Errorf("could not create build struct: %w", err)
  }

  // Loop through Builds and get more detailed information
  for _, v := range builds {
    err = getBuildInfo(v, repo.Namespace, repo.Name)
    if err != nil {
      return fmt.Errorf("could not get build info: %w", err)
    }
  }

  return nil
}

func getBuildInfo(build Build, repoNamespace string, repoName string) error{
  // Set empty BuildInfo Struct
  var buildInfo BuildInfo

  // Set build variables
  build.Time = time.Unix(build.Started, 0)
  build.RepoTeam = repoNamespace
  build.RepoName = repoName

  // Do API Call to Drone
  var subUrlPath = fmt.Sprintf("/api/repos/%s/%s/builds/%s", repoNamespace, repoName, strconv.Itoa(build.Number))
  data, err := api.ApiRequest(subUrlPath)
  if err != nil {
    return fmt.Errorf("error getting buildinfo: %w", err)
  }

  // Create struct of API Request data
  if err := json.Unmarshal([]byte(data), &buildInfo); err != nil {
    return fmt.Errorf("could not create buildinfo struct: %w", err)
  }

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
      build.Pipeline = y.Name
      influxdb.Run(structs.Map(build), y.Name)
    }
  }

  return nil
}