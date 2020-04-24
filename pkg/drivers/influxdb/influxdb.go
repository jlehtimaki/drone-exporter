package influxdb

import (
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/jlehtimaki/drone-exporter/pkg/env"
)

var (
	influxAddress = env.GetEnv("DB_ADDRESS", "http://localhost:8086")
	database      = env.GetEnv("DATABASE", "example")
	username      = env.GetEnv("DB_USERNAME", "foo")
	password      = env.GetEnv("DB_PASSWORD", "bar")
	influxClient  client.Client
)

func GetClient() (client.Client, error) {
	if influxClient != nil {
		return influxClient, nil
	}

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     influxAddress,
		Username: username,
		Password: password,
	})

	if err != nil {
		return nil, err
	}

	influxClient = c
	return influxClient, nil
}

func Close() error {
	c, err := GetClient()
	if err != nil {
		return err
	}

	if c == nil {
		return nil
	}

	return c.Close()
}

func Run(builds map[string]interface{}, pipelineName string) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	tags := map[string]string{"Pipeline": pipelineName}
	// Create a point and add to batch
	pt, err := client.NewPoint("drone", tags, builds, builds["Time"].(time.Time))
	if err != nil {
		return err
	}
	bp.AddPoint(pt)

	// Write the batch
	c, err := GetClient()
	if err != nil {
		return err
	}
	if err := c.Write(bp); err != nil {
		return err
	}

	return nil
}

func Batch(measurement string, fieldList []map[string]interface{}) error {
	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  database,
		Precision: "s",
	})
	if err != nil {
		return err
	}

	for _, fields := range fieldList {
		// Create a point and add to batch
		tags := fields["Tags"].(map[string]string)
		delete(fields, "Tags")
		pt, err := client.NewPoint(measurement, tags, fields, fields["Time"].(time.Time))
		if err != nil {
			return err
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	c, err := GetClient()
	if err != nil {
		return err
	}
	if err := c.Write(bp); err != nil {
		return err
	}

	return nil
}
