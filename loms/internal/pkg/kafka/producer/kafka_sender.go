package producer

import (
	"github.com/IBM/sarama"
)

type KafkaSender struct {
	TopicName string
	Producer  sarama.SyncProducer
}

func NewKafkaSender(topicName string, producer sarama.SyncProducer) KafkaSender {
	return KafkaSender{
		TopicName: topicName,
		Producer:  producer,
	}
}
