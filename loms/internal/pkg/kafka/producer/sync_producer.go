package producer

import (
	"fmt"

	"github.com/IBM/sarama"
	"route256/loms/internal/pkg/kafka"
)

// NewSyncProducer creates a sarama.SyncProducer for the given brokers.
func NewSyncProducer(conf kafka.Config, opts ...Option) (sarama.SyncProducer, error) {
	config, err := PrepareConfig(opts...)
	if err != nil {
		return nil, fmt.Errorf("prepare config: %w", err)
	}

	syncProducer, err := sarama.NewSyncProducer(conf.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("NewSyncProducer failed: %w", err)
	}

	return syncProducer, nil
}
