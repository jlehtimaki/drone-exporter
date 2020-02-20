package api

import (
  "fmt"
  "github.com/jlehtimaki/drone-exporter/pkg/env"
  "io/ioutil"
  "net/http"
)

func ApiRequest(subUrlPath string) (string, error) {

  droneUrl := env.GetEnv("DRONE_URL","")
  token := env.GetEnv("TOKEN","")

  if droneUrl == "" || token == "" {
    return "", fmt.Errorf("could not read DRONE_URL or TOKEN from env")
  }

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