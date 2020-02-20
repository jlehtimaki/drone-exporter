package influxdb

import (
	"context"
	"github.com/influxdata/influxdb-client-go"
	"log"
	"time"
)

var (
	myHTTPInfluxAddress = "foobar"
	myToken = "foobar"
)

func Run(){
	influx, err := influxdb.New(myHTTPInfluxAddress, myToken, influxdb.WithHTTPClient(myHTTPClient))
	if err != nil {
		panic(err) // error handling here; normally we wouldn't use fmt but it works for the example
	}

	// we use client.NewRowMetric for the example because it's easy, but if you need extra performance
	// it is fine to manually build the []client.Metric{}.
	myMetrics := []influxdb.Metric{
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 8, time.UTC)),
		influxdb.NewRowMetric(
			map[string]interface{}{"memory": 1000, "cpu": 0.93},
			"system-metrics",
			map[string]string{"hostname": "hal9000"},
			time.Date(2018, 3, 4, 5, 6, 7, 9, time.UTC)),
	}

	// The actual write..., this method can be called concurrently.
	if _, err := influx.Write(context.Background(), "my-awesome-bucket", "my-very-awesome-org", myMetrics...); err != nil {
		log.Fatal(err) // as above use your own error handling here.
	}
	influx.Close() // closes the client.  After this the client is useless.
}