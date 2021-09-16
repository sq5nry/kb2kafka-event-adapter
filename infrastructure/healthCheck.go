package infrastructure

import (
	"fmt"
	"github.com/golang/glog"
	"io/ioutil"
	"kb2kafka-event-adapter/config/listener"
	"net/http"
	"strconv"
)

func OnHealthCheck(writer http.ResponseWriter, config *listener.ListenerConfiguration) {
	glog.Info("healthcheck requested")
	brokerStatus := checkHealth(writer, config.KafkaBrokerHealthCheckUrl)
	clusterStatus := checkHealth(writer, config.KafkaClusterHealthCheckUrl)
	fmt.Fprintf(writer, "broker=%s, cluster=%s", brokerStatus, clusterStatus)
}

func checkHealth(writer http.ResponseWriter, url string) string {
	msg, status, err := isHealthy(url)
	if err != nil || !status {
		sendErrorResponse(writer, http.StatusInternalServerError, "http protocol error in health check: " + msg, err)
	}
	glog.Infof("health check response from %s: %s, %s", url, msg, err)
	return msg
}

func isHealthy(healthCheckUrl string) (responseMsg string, isSuccessful bool, error error) {
	resp, err := http.Get(healthCheckUrl)
	if err != nil {
		return "", false, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body), false, err
	}
	return strconv.Itoa(resp.StatusCode) + "/" + string(body), resp.StatusCode == http.StatusOK, nil
}

func sendErrorResponse(writer http.ResponseWriter, code int, description string, err error) {
	msg := fmt.Sprintf(description+": %v", err)
	glog.Error(msg)
	http.Error(writer, msg, code)
}
