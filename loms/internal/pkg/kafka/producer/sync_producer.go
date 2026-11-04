package producer

import (
	"fmt"

	"github.com/IBM/sarama"
	"route256/loms/internal/pkg/kafka"
)

func NewSyncProducer(conf kafka.Config, opts ...Option) (sarama.SyncProducer, error) {
	config := PrepareConfig(opts...)

	syncProducer, err := sarama.NewSyncProducer(conf.Brokers, config)
	if err != nil {
		return nil, fmt.Errorf("NewSyncProducer failed: %w", err)
	}

	return syncProducer, nil
}
