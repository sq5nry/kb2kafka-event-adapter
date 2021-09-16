package repository

import (
	"errors"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	kafkaCfg "kb2kafka-event-adapter/config/kafka"
)

func NewProducer(config *kafkaCfg.KafkaConfiguration) *kafka.Producer {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServer,
		"client.id":         config.ClientId,
	})
	if err != nil {
		glog.Fatal(err)
	}
	glog.Infof("kafka producer created: %v", kafkaProducer)
	return kafkaProducer
}

func Dispatch(message []byte, key string, producer *kafka.Producer, configuration *kafkaCfg.KafkaConfiguration) error {
	deliveryChan := make(chan kafka.Event)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &configuration.ConsumerTopicName, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            []byte(key),
	}, deliveryChan)
	if err != nil {
		glog.Warningf("creating a message for delivery failed: %#v\n", err)
		return err
	}

	event := <-deliveryChan
	ack := event.(*kafka.Message)
	if ack.TopicPartition.Error != nil {
		glog.Warning("delivery failed: ", ack.TopicPartition.Error)
		return errors.New("delivery failed")
	} else {
		glog.Infof("delivered message [%s] with key [%s] to topic %s [%d] at offset %v\n",
			message, key, *ack.TopicPartition.Topic, ack.TopicPartition.Partition, ack.TopicPartition.Offset)
		return nil
	}
}
