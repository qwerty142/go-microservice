package producer

import (
	"time"

	"github.com/IBM/sarama"
)

func PrepareConfig(opts ...Option) *sarama.Config {
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
		_ = opt.Apply(c)
	}

	return c
}
