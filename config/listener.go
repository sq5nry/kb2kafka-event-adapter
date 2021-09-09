package config

import (
	"github.com/magiconair/properties"
	"log"
)

type ListenerConfiguration struct {
	ListenerAddress string `properties: "listenerAddress"`
	EndpointPath string `properties: "endpointPath"`
}

func GetListenerConfig() ListenerConfiguration  {
	configuration := properties.MustLoadFile("./resources/kbEventAdapterConfig.properties", properties.UTF8)

	var listenerConfiguration ListenerConfiguration
	if err := configuration.Decode(&listenerConfiguration); err != nil {
		log.Fatal(err)
	}

	log.Printf("DEBUG: loaded http listener configuration: %s", configuration.String())

	return listenerConfiguration
	//if host = os.Getenv("TODO_EP_HOST"); len(host) == 0 {
	//	host = "localhost"
	//}
	//if port = os.Getenv("TODO_EP_POST"); len(port) == 0 {
	//	port = "8080"
	//}
	//if endpointPath = os.Getenv("TODO_EP_PATH"); len(endpointPath) == 0 {
	//	endpointPath = "endpoint"
	//}
}
