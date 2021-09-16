package confirmation

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	kafkaCfg "kb2kafka-event-adapter/config/kafka"
	"kb2kafka-event-adapter/domain/entity"
	"time"
)

func NewConfirmationConsumer(config *kafkaCfg.KafkaConfiguration) *kafka.Consumer {
	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServer,
		"group.id":          config.GroupId,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		panic(err)
	}

	subscriberError := kafkaConsumer.Subscribe(config.ResponseTopicName, nil)
	if subscriberError != nil {
		panic(subscriberError)
	}

	glog.Infof("kafka consumer created: %v", kafkaConsumer)
	return kafkaConsumer
}

func GetMessage(kafkaConsumer *kafka.Consumer, timeout time.Duration) (*ConfirmationMessage, error) {
	rawCfmMessage, err := kafkaConsumer.ReadMessage(timeout)

	if err == nil {
		glog.Infof("received confirmation message from kafka %s", string(rawCfmMessage.Value))
		cfmMessage, err := CreateConfirmationMessageFromJSON(rawCfmMessage.Value)
		return cfmMessage, err
	} else {
		glog.Error("could not get confirmation message from the persistence")
		return &ConfirmationMessage{}, err
	}
}

func ValidateConfirmationMessage(cfmMessage *ConfirmationMessage, ratingMsg *entity.RatingMessage) bool {
	return cfmMessage.Id == ratingMsg.Id;
}