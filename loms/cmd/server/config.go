package main

import (
	"route256/loms/internal/pkg/kafka"
)

type ProducerConfig struct {
	TopicName string
}

type Config struct {
	kafka    kafka.Config
	producer ProducerConfig
}

func GetConfig(f flags) Config {
	return Config{
		kafka: kafka.Config{
			Brokers: []string{
				f.bootstrapServer,
			},
		},
		producer: ProducerConfig{
			TopicName: f.topic,
		},
	}
}
