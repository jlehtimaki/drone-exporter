package main

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    //"github.com/jlehtimaki/drone_exporter@initial/cmd/drivers/influxdb"
)


var (
    droneUrl	= "https://drone.tools.vrk-kubernetes.net"
    token		= "8E25GPQgPqaaJkTgOBDiQcJYkiViGX6W"
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

func apiRequest(subUrlPath string) (string, error) {
    urlPath := fmt.Sprintf( "%s%s", droneUrl, subUrlPath)
    client := &http.Client{}
    req, err := http.NewRequest("GET", urlPath, nil)
    bearerToken := fmt.Sprintf("Bearer %s", token)
    req.Header.Add("Authorization", bearerToken)
    resp, err := client.Do(req)

    defer resp.Body.Close()
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", fmt.Errorf("could not read response: %w", err)
    }
    return string(body), nil
}

func getRepos() error {
    var subUrlPath = "/api/user/repos"
    data, err := apiRequest(subUrlPath)
    if err != nil {
        return fmt.Errorf("could not get repos: %w", err)
    }

    var repos []Repo
    if err := json.Unmarshal([]byte(data), &repos); err != nil {
        return fmt.Errorf( "could not create repos struct: %w", err)
    }
    for _, v := range repos {
    	if v.Active {
    		err := getBuilds(v)
    		if err != nil {
    			return err
    		}
    	}
    }
    return nil
}

func getBuilds(repos Repo) error {
    subUrlPath := fmt.Sprintf("/api/repos/%s" +
        "/%s/builds", repos.Namespace, repos.Name)

    data, err := apiRequest(subUrlPath)
    if err != nil {
        return fmt.Errorf("could not get builds: %w", err)
    }

    var builds []Build
    if err:= json.Unmarshal([]byte(data), &builds); err != nil {
        return fmt.Errorf("could not create build struct: %w", err)
    }

    for _, v := range builds {
        if v.Status == "success" {
            fmt.Println(v.Message)
        }
    }

    return nil
}

func main()  {
    err := getRepos()
    if err != nil {
        fmt.Printf("error: %s", err)
    }
}