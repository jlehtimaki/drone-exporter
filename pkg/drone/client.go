package drone

import (
	"context"

	"github.com/drone/drone-go/drone"
	dronecli "github.com/drone/drone-go/drone"
	"github.com/jlehtimaki/drone-exporter/pkg/env"
	"golang.org/x/oauth2"
)

var (
	host   = env.GetEnv("DRONE_URL", "")
	token  = env.GetEnv("DRONE_TOKEN", "")
	client *drone.Client
)

func GetClient() *dronecli.Client {
	if client != nil {
		return client
	}

	if host == "" || token == "" {
		return nil
	}

	config := new(oauth2.Config)
	auth := config.Client(
		context.TODO(),
		&oauth2.Token{
			AccessToken: token,
		},
	)

	c := dronecli.NewClient(host, auth)
	client = &c

	return client
}

func GetHost() string {
	return host
}

//func ApiRequest(subUrlPath string) (string, error) {
//	c := getClient()
//	if c == nil {
//		return "", errors.New("unable to create drone client, please check env DRONE_URL and DRONE_TOKEN")
//	}
//	urlPath := fmt.Sprintf("%s%s", droneUrl, subUrlPath)
//	client := &http.Client{}
//	req, err := http.NewRequest("GET", urlPath, nil)
//	if err != nil {
//		return "", fmt.Errorf("error creating request: %w", err)
//	}
//	bearerToken := fmt.Sprintf("Bearer %s", token)
//	req.Header.Add("Authorization", bearerToken)
//	resp, err := client.Do(req)
//	if err != nil {
//		return "", fmt.Errorf("error doing request: %w", err)
//	}
//	defer resp.Body.Close()
//	body, err := ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return "", fmt.Errorf("could not read response: %w", err)
//	}
//	return string(body), nil
//}
