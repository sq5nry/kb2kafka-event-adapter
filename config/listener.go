package config

import "os"

func GetListenerConfig() (string, string, string) {
	var host, port, endpointPath string

	if host = os.Getenv("TODO_EP_HOST"); len(host) == 0 {
		host = "localhost"
	}
	if port = os.Getenv("TODO_EP_POST"); len(port) == 0 {
		port = "8080"
	}
	if endpointPath = os.Getenv("TODO_EP_PATH"); len(endpointPath) == 0 {
		endpointPath = "endpoint"
	}

	return host, port, endpointPath
}
