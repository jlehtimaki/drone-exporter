package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/jlehtimaki/drone-exporter/pkg/env"
)

var (
	droneUrl = env.GetEnv("DRONE_URL", "")
	token    = env.GetEnv("TOKEN", "")
)

func ApiRequest(subUrlPath string) (string, error) {
	if droneUrl == "" || token == "" {
		return "", fmt.Errorf("could not read DRONE_URL or TOKEN from env")
	}

	urlPath := fmt.Sprintf("%s%s", droneUrl, subUrlPath)
	client := &http.Client{}
	req, err := http.NewRequest("GET", urlPath, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}
	bearerToken := fmt.Sprintf("Bearer %s", token)
	req.Header.Add("Authorization", bearerToken)
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error doing request: %w", err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("could not read response: %w", err)
	}
	return string(body), nil
}
