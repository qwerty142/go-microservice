package main

import (
	"route256/notifier/internal/pkg/kafka"
)

type config struct {
	KafkaConfig kafka.Config
}

func newConfig(f flags) config {
	return config{
		KafkaConfig: kafka.Config{
			Brokers: []string{
				f.bootstrapServer,
			},
		},
	}
}
