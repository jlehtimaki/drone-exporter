package main

import (
  "encoding/json"
  "fmt"
  "github.com/jlehtimaki/drone-exporter/cmd/api"
  //"github.com/jlehtimaki/drone_exporter@initial/cmd/drivers/influxdb"
)

type Repo struct {
  Id          int     `json:"Id"`
  Name        string	`json:"Name"`
  Active      bool	`json:"Active"`
  Namespace   string  `json:"Namespace"`
}

type Build struct {
  Id          int     `json:"Id"`
  Trigger     string  `json:"Trigger"`
  Status      string  `json:"Status"`
  Number      int     `json:"Number"`
  Event       string  `json:"Event"`
  Action      string  `json:"Action"`
  Link        string  `json:"Link"`
  Message     string  `json:"Message"`
  Ref         string  `json:"Ref"`
  Source      string  `json:"Source"`
  Target      string  `json:"Target"`
  Sender      string  `json:"Sender"`
  Started     int     `json:"Started"`
  Finished    int     `json:"Finished"`
}

type RepoWithBuilds struct {
  Repo    Repo
  Builds   []Build
}

func (rwb *RepoWithBuilds) AddBuild(build Build){
  rwb.Builds = append(rwb.Builds, build)
}

func getRepos() error {
  var subUrlPath = "/api/user/repos"
  data, err := api.apiRequest(subUrlPath)
  if err != nil {
    return fmt.Errorf("could not get repos: %w", err)
  }

  var repos []Repo
  var repoWithBuilds []RepoWithBuilds
  if err := json.Unmarshal([]byte(data), &repos); err != nil {
    return fmt.Errorf( "could not create repos struct: %w", err)
  }
  for _, v := range repos {
    if v.Active {
      rbw := RepoWithBuilds{}
      rbw.Repo = v
      rbw, err := getBuilds(v, rbw)
      if err != nil {
        return err
      }
      repoWithBuilds = append(repoWithBuilds, rbw)
    }
  }

  return nil
}

func getBuilds(repos Repo, rbw RepoWithBuilds) (RepoWithBuilds, error) {
  subUrlPath := fmt.Sprintf("/api/repos/%s" +
    "/%s/builds", repos.Namespace, repos.Name)

  data, err := api.apiRequest(subUrlPath)
  if err != nil {
    return rbw, fmt.Errorf("could not get builds: %w", err)
  }

  var builds []Build
  if err:= json.Unmarshal([]byte(data), &builds); err != nil {
    return rbw, fmt.Errorf("could not create build struct: %w", err)
  }

  for _, v := range builds {
    rbw.AddBuild(v)
  }

  return rbw, nil
}

func main()  {
  err := getRepos()
  if err != nil {
    fmt.Printf("error: %s", err)
  }
}