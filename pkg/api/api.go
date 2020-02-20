package api

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "os"
)

func ApiRequest(subUrlPath string) (string, error) {

  droneUrl := os.Getenv("DRONE_URL")
  token := os.Getenv("TOKEN")

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
