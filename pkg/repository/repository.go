package repository

import (
  "encoding/json"
  "fmt"
  "github.com/fatih/structs"
  "github.com/jlehtimaki/drone-exporter/pkg/api"
  "github.com/jlehtimaki/drone-exporter/pkg/drivers"
  "time"
)

type Repo struct {
  Id          int     `json:"Id"`
  Name        string	`json:"Name"`
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
}

type RepoWithBuilds struct {
  Repo    Repo
  Builds   []Build
}

func (rwb *RepoWithBuilds) AddBuild(build Build){
  rwb.Builds = append(rwb.Builds, build)
}


func GetRepos() error {
  var subUrlPath = "/api/user/repos"
  data, err := api.ApiRequest(subUrlPath)
  if err != nil {
    return fmt.Errorf("error getting repositories: %w", err)
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
  for _, v := range repoWithBuilds{
    for _, x := range v.Builds {
      x.RepoName = v.Repo.Name
      x.Time = time.Unix(x.Started, 0)
      influxdb.Run(structs.Map(x))
    }
  }
  return nil
}

func getBuilds(repos Repo, rbw RepoWithBuilds) (RepoWithBuilds, error) {
  subUrlPath := fmt.Sprintf("/api/repos/%s" +
    "/%s/builds", repos.Namespace, repos.Name)

  data, err := api.ApiRequest(subUrlPath)
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
