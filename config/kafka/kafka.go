package kafka

import (
	"errors"
	"github.com/golang/glog"
	"os"
	"reflect"
)

const TAG_ENV_VAR = "env_var"

type KafkaConfiguration struct {
	ConsumerTopicName string `env_var:"KAFKA_TOPIC_NAME"`			//repository topic for KB events
	BootstrapServer   string `env_var:"KAFKA_BOOTSTRAP_SERVER"`
	GroupId           string `env_var:"KAFKA_GROUP_ID"`
	ResponseTopicName string `env_var:"KAFKA_RESPONSE_TOPIC_NAME"`	//consumer topic for confirmation events from the persistence reader
	ClientId		  string
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

	configureClientId(configuration)

	return configuration, nil
}

func configureClientId(configuration *KafkaConfiguration) {
	host, err := os.Hostname()
	if err != nil {
		glog.Fatal(err.Error())
	}
	configuration.ClientId = host
}

func readEnvVar(configElems reflect.Value, paramOffset int, configuration *KafkaConfiguration, configElem reflect.Value) (*KafkaConfiguration, error) {
	name := configElems.Type().Field(paramOffset).Name
	field, _ := reflect.TypeOf(configuration).Elem().FieldByName(name)

	envVarName := field.Tag.Get(TAG_ENV_VAR)
	if len(envVarName) > 0 {
		if envVal, found := os.LookupEnv(envVarName); found {
			glog.Infof("initializing kafka configuration from env variable=%s, value=%s\n", envVarName, envVal)
			configElem.SetString(envVal)
		} else {
			return configuration, errors.New(envVarName + " env var not found")
		}
	}
	return nil, nil
}
