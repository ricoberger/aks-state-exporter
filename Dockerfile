FROM golang:1.23.0 as build
WORKDIR /aks-state-exporter
COPY go.mod go.sum /aks-state-exporter/
RUN go mod download
COPY . .
RUN export CGO_ENABLED=0 && make build

FROM alpine:3.20.2
RUN apk update && apk add --no-cache ca-certificates
RUN mkdir /aks-state-exporter
COPY --from=build /aks-state-exporter/bin/aks-state-exporter /aks-state-exporter
WORKDIR /aks-state-exporter
USER nobody
ENTRYPOINT  [ "/aks-state-exporter/aks-state-exporter", "start" ]
