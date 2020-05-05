package influxdb

import (
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/jlehtimaki/drone-exporter/pkg/env"
	log "github.com/sirupsen/logrus"
	"time"
)


func Run(builds map[string]interface{}, pipelineName string){
	influxAddress := env.GetEnv("DB_ADDRESS", "http://localhost:8086")
	database      := env.GetEnv("DATABASE", "example")
	username      := env.GetEnv("DB_USERNAME", "foo")
	password      := env.GetEnv("DB_PASSWORD", "bar")

	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr: influxAddress,
		Username: username,
		Password: password,
	})
	if err != nil {
		log.Fatal("Error creating InfluxDB Client: ", err.Error())
	}

		// Create a new point batch
		bp, err := client.NewBatchPoints(client.BatchPointsConfig{
			Database:  database,
			Precision: "ms",
		})
		if err != nil {
			log.Fatal(err)
		}

		tags := map[string]string{"Pipeline": pipelineName}
		// Create a point and add to batch
		pt, err := client.NewPoint("drone", tags, builds, builds["Time"].(time.Time))
		if err != nil {
			log.Fatal(err)
		}
		bp.AddPoint(pt)

		// Write the batch
		if err := c.Write(bp); err != nil {
			log.Fatal(err)
		}

	// Close client resources
	if err := c.Close(); err != nil {
		log.Fatal(err)
	}

}