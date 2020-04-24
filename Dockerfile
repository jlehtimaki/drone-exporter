FROM golang:alpine AS builder
ENV GOARCH=amd64 GOOS=linux CGO_ENABLED=0
WORKDIR /build

COPY . .

RUN go build

FROM golang:alpine

COPY --from=builder /build/drone-exporter /bin/

ENTRYPOINT ["/bin/drone-exporter"]
