package kafka

import (
	"errors"
	"kb2kafka-event-adapter/config"
	"log"
	"os"
	"reflect"
)

type KafkaConfiguration struct {
	ConsumerTopicName string `env_var:"KAFKA_TOPIC_NAME"`
	BootstrapServer   string `env_var:"KAFKA_BOOTSTRAP_SERVER"`
	GroupId           string `env_var:"KAFKA_GROUP_ID"`
}

func GetKafkaConfiguration() (*KafkaConfiguration, error) {
	configuration := new(KafkaConfiguration)
	var configElems = reflect.ValueOf(configuration).Elem()
	for i := 0; i < configElems.NumField(); i++ {
		configElem := configElems.Field(i)
		if configElem.Type().Kind() == reflect.String {
			kafkaConfiguration, err := readEnvVar(configElems, i, configuration, configElem)
			if err != nil {
				return kafkaConfiguration, err
			}
		}
	}

	return configuration, nil
}

func readEnvVar(configElems reflect.Value, paramOffset int, configuration *KafkaConfiguration, configElem reflect.Value) (*KafkaConfiguration, error) {
	name := configElems.Type().Field(paramOffset).Name
	field, _ := reflect.TypeOf(configuration).Elem().FieldByName(name)

	envVarName := field.Tag.Get(config.TAG_ENV_VAR)
	if envVal, found := os.LookupEnv(envVarName); found {
		log.Printf("DEBUG: initializing kafka configuration from env variable=%s, value=%s\n", envVarName, envVal)
		configElem.SetString(envVal)
	} else {
		return configuration, errors.New(envVarName + " env var not found")
	}
	return nil, nil
}
