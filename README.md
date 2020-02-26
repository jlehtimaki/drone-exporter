# drone_exporter
Drone exporter is meant to push Drone build data to DB like InfluxDB\
From different databases you can then show your build information e.g. in Grafana

## Supported Databases
- InfluxDB

## Build
```bash
go build .cmd/drone-exporter
```

## Docker
```bash
docker built -t lehtux/drone-exporter .
```

## Parameters
| Name          | Description               | Default               | Required  |
| ------------- |:-------------------------:| ---------------------:|:---------:|
| INTERVAL      | Time between runs         | 2                     | No        |
| DRONE_URL     | Drone URL                 | NIL                   | Yes       |
| TOKEN         | Drone Token               | NIL                   | Yes       |
| DB_ADDRESS    | Database address          | http://localhost:8086 | No        |
| DB_USERNAME   | Database username         | foo                   | No        |
| DB_PASSWORD   | Database password         | bar                   | No        |
| DATABASE      | Database name             | example               | No        |


### Database schema
This table shows the current information that gets stored to the database

| Name          | Description               | Type      |
| ------------- |:-------------------------:| ---------:|
| Id            | ID of the build           | int       |
| Trigger       | Trigger of the build      | string    |
| Status        | Status of the build       | string    |
| Number        | Build number              | int       |
| Event         | Event that made trigger   | string    |
| Action        | Action of the build       | string    |
| Link          | Link to the commit        | string    |
| Message       | Commit message            | string    |
| Source        | Source branch             | string    |
| Target        | Target branch             | string    |
| Sender        | Person who made the event | string    |
| Started       | Start time                | int64     |
| Finished      | Finish time               | int64     |
| Time          | Time in non unix epoch    | Time      |
| RepoName      | Repository name           | string    |
| RepoTeam      | Team name                 | string    |
| Pipeline      | Pipeline name             | string    |
| BuilState     | Build status in integer   | int       |
