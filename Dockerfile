FROM golang:1-alpine AS builder
WORKDIR /build
COPY . .
RUN go build

FROM alpine
COPY --from=builder /build/drone-exporter /bin/

ENTRYPOINT ["/bin/drone-exporter"]
