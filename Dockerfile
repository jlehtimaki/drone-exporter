FROM golang:alpine AS builder
ENV GOARCH=amd64 GOOS=linux CGO_ENABLED=0
WORKDIR /build

COPY . .

RUN cd cmd/drone-exporter && go build

FROM golang:alpine

COPY --from=builder /build/cmd/drone-exporter/drone-exporter /bin/

ENTRYPOINT ["/bin/drone-exporter"]
