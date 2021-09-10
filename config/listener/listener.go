package listener

import (
	"errors"
	"kb2kafka-event-adapter/config"
	"log"
	"os"
	"reflect"
)

type ListenerConfiguration struct {
	ListenerAddress string `env_var:"LISTENER_ADDRESS"` // :8082
	EndpointPath    string `env_var:"LISTENER_PATH"`    // /callmeback
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

	envVarName := field.Tag.Get(config.TAG_ENV_VAR)
	if envVal, found := os.LookupEnv(envVarName); found {
		log.Printf("DEBUG: initializing listener configuration from env variable=%s, value=%s\n", envVarName, envVal)
		configElem.SetString(envVal)
	} else {
		return configuration, errors.New(envVarName + " env var not found")
	}
	return nil, nil
}
