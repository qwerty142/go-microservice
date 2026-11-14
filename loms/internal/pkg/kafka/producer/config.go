package producer

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

// PrepareConfig builds the base sarama config for a sync producer and applies the given options.
func PrepareConfig(opts ...Option) (*sarama.Config, error) {
	c := sarama.NewConfig()

	c.Producer.Partitioner = sarama.NewHashPartitioner
	c.Producer.RequiredAcks = sarama.WaitForAll
	c.Producer.Idempotent = true

	c.Producer.Retry.Max = 100
	c.Producer.Retry.Backoff = 5 * time.Millisecond

	c.Net.MaxOpenRequests = 1

	c.Producer.CompressionLevel = sarama.CompressionLevelDefault
	c.Producer.Compression = sarama.CompressionNone

	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = true

	for _, opt := range opts {
		if err := opt.Apply(c); err != nil {
			return nil, fmt.Errorf("apply option: %w", err)
		}
	}

	return c, nil
}
