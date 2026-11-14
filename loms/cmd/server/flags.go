package main

import (
	"flag"
)

type flags struct {
	topic           string
	bootstrapServer string
	jaegerURL       string
}

var cliFlags = flags{}

func init() {
	flag.StringVar(&cliFlags.topic, "topic", "loms.order-events", "topic to produce")
	flag.StringVar(&cliFlags.bootstrapServer, "bootstrap-server", "kafka0:29092", "kafka broker host and port")
	flag.StringVar(&cliFlags.jaegerURL, "jaeger-url", "http://jaeger:4318", "jaeger OTLP HTTP endpoint")

	flag.Parse()
}
