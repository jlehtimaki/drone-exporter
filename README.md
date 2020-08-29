# Drone Exporter
Daemon to extract data from the Drone API and push it into a backend powered by a database driver.

## Key Features
* Supports Multi-threaded, process 1 repo's builds per thread
* On first boot, will grab all data from Drone and import it
* Queries a repo's last build number from the database and skips processing if there are no new builds
* Calculates build's WaitTime and Duration for easy charting.
* Never imports data from running jobs to avoid erroneous WaitTime/Duration data
* Data points for builds, stages and steps for granular charts

## Supported Database Drivers
List of supported backends for Grafana and which ones Drone Exporter currently supports.
- [ ] Prometheus
- [ ] Graphite
- [ ] OpenTSDB
- [x] InfluxDB
- [ ] Loki
- [ ] Elasticsearch
- [ ] MySQL
- [ ] PostgreSQL
- [ ] MSSQL Server

### Schema

#### Measurement: builds 
```go
&types.Build{
    Time:     time.Unix(buildInfo.Started, 0),
    Number:   buildInfo.Number,
    WaitTime: buildInfo.Started - buildInfo.Created,
    Duration: buildInfo.Finished - buildInfo.Started,
    Source:   buildInfo.Source,
    Target:   buildInfo.Target,
    Started:  buildInfo.Started,
    Created:  buildInfo.Created,
    Finished: buildInfo.Finished,
    BuildId:  build.Number,
    Tags: map[string]string{
        "Slug":    repo.Slug,
        "BuildId": fmt.Sprintf("build-%d", buildInfo.Number),
    },
}
```

#### Measurement: stages
```go
&types.Stage{
    Time:     time.Unix(stage.Started, 0),
    WaitTime: stage.Started - stage.Created,
    Duration: stage.Stopped - stage.Started,
    OS:       stage.OS,
    Arch:     stage.Arch,
    Status:   stage.Status,
    Name:     stage.Name,
    BuildId:  build.Number,
    Tags: map[string]string{
        "Slug":    repo.Slug,
        "BuildId": fmt.Sprintf("build-%d", build.Number),
        "Sender":  build.Sender,
        "Name":    stage.Name,
        "OS":      stage.OS,
        "Arch":    stage.Arch,
        "Status":  stage.Status,
    },
}
```

#### Measurement steps
```go
&types.Step{
    Time:     time.Unix(step.Started, 0),
    Duration: step.Stopped - step.Started,
    Name:     step.Name,
    Status:   step.Status,
    Tags: map[string]string{
        "Slug":    repo.Slug,
        "BuildId": fmt.Sprintf("build-%d", build.Number),
        "Sender":  build.Sender,
        "Name":    step.Name,
        "Status":  step.Status,
    },
}
```

## Build
### From Source
```bash
go build -mod vendor
DRONE_URL=https://dronezerver.xyz DRONE_TOKEN=abcde12345 ./drone-exporter
```

### Docker
```bash
docker build -t lehtux/drone-exporter .
# add more envs using -e, see below
docker run -d -e DRONE_URL https://droneserver.xyz -e DRONE_TOKEN abcde12345 lehtux/drone-exporter
```

## Environment Variables
| Name          | Description               | Default               | Required  |
| ------------- |:-------------------------:| ---------------------:|:---------:|
| INTERVAL      | Time between runs in minutes | 2                  | No        |
| THREADS       | Number of repos to process simultaneously | 10     | No        |
| DRONE_URL     | Drone URL                 | NIL                   | Yes       |
| DRONE_TOKEN   | Drone Token               | NIL                   | Yes       |
| INFLUXDB_ADDRESS | Database address       | http://influxdb:8086 | No        |
| INFLUXDB_USERNAME   | Database username   | foo                   | No        |
| INFLUXDB_PASSWORD   | Database password   | bar                   | No        |
| INFLUXDB_DATABASE   | Database name       | example               | No        |

