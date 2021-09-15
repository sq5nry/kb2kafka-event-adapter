package config

import (
	"github.com/golang/glog"
	"kb2kafka-event-adapter/config/kafka"
	"kb2kafka-event-adapter/config/listener"
)

type Configuration struct {
	ListenerConfiguration *listener.ListenerConfiguration
	KafkaConfiguration    *kafka.KafkaConfiguration
}

func GetConfiguration() *Configuration {
	kafkaConfig, err := kafka.GetKafkaConfiguration()
	if err != nil {
		glog.Fatalf("error in kafka configuration: %v", err)
	}

	listenerConfig, err := listener.GetListenerConfig()
	if err != nil {
		glog.Fatalf("error in listener configuration: %v", err)
	}

	return &Configuration{
		ListenerConfiguration: listenerConfig,
		KafkaConfiguration:    kafkaConfig,
	}
}
