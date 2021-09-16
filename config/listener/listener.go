package listener

import (
	"errors"
	"github.com/golang/glog"
	"os"
	"reflect"
)

const TAG_ENV_VAR = "env_var"

type ListenerConfiguration struct {
	ListenerAddress            string `env_var:"LISTENER_ADDRESS"`
	EndpointPath               string `env_var:"LISTENER_PATH"`
	KafkaBrokerHealthCheckUrl  string `env_var:"KAFKA_BROKER_HEALTHCHECK_URL"`
	KafkaClusterHealthCheckUrl string `env_var:"KAFKA_CLUSTER_HEALTHCHECK_URL"`
}

func GetListenerConfig() (*ListenerConfiguration, error) {
	configuration := new(ListenerConfiguration)
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

func readEnvVar(configElems reflect.Value, paramOffset int, configuration *ListenerConfiguration, configElem reflect.Value) (*ListenerConfiguration, error) {
	name := configElems.Type().Field(paramOffset).Name
	field, _ := reflect.TypeOf(configuration).Elem().FieldByName(name)

	envVarName := field.Tag.Get(TAG_ENV_VAR)
	if len(envVarName) > 0 {
		if envVal, found := os.LookupEnv(envVarName); found {
			glog.Infof("initializing listener configuration from env variable=%s, value=%s\n", envVarName, envVal)
			configElem.SetString(envVal)
		} else {
			return configuration, errors.New(envVarName + " env var not found")
		}
	}
	return nil, nil
}
