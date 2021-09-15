package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	"kb2kafka-event-adapter/config"
	"kb2kafka-event-adapter/domain/broker"
	"kb2kafka-event-adapter/domain/entity"
	kafkaRepo "kb2kafka-event-adapter/domain/repository"
	"net/http"
)

var configuration *config.Configuration
var kafkaProducer *kafka.Producer

func init() {
	flag.Parse() //workaround for glog "logging before flag.Parse" issue
	configuration = config.GetConfiguration()
	kafkaProducer = kafkaRepo.NewProducer(configuration.KafkaConfiguration)
}

//TODO readiness & liveness probing
func main() {
	http.HandleFunc(configuration.ListenerConfiguration.EndpointPath, onHttpReq)

	glog.Info("starting listener")
	if err := http.ListenAndServe(configuration.ListenerConfiguration.ListenerAddress, nil); err != nil {
		glog.Exit(err.Error())
	}
}

func onHttpReq(writer http.ResponseWriter, request *http.Request) {
	if ensureHttpMethod(writer, request) { return }

	if kbEvent, err := extractKbEvent(request); err != nil {
		handleError(writer, http.StatusBadRequest, "decoding error", err)
	} else {
		glog.Info("KbEvent received", kbEvent)
		storeKbEvent(writer, kbEvent)
	}
}

func extractKbEvent(request *http.Request) (*broker.KBEvent, error) {
	var kbEvent broker.KBEvent
	err := json.NewDecoder(request.Body).Decode(&kbEvent)
	return &kbEvent, err
}

func storeKbEvent(writer http.ResponseWriter, kbEvent *broker.KBEvent) {
	ratingMsg := entity.CreateRatingFromKbEvent(kbEvent)
	ratingMsgBytes, err := json.Marshal(ratingMsg)
	if err == nil {
		glog.Infof("pushing KB request to kafka, id=%s", ratingMsg.Id)
		err = kafkaRepo.Dispatch(ratingMsgBytes, ratingMsg.Id, kafkaProducer, configuration.KafkaConfiguration)
		if err != nil {
			handleError(writer, http.StatusInternalServerError, "could not push to kafka", err)
		}
	} else {
		handleError(writer, http.StatusInternalServerError, "could not create a rating message", err)
	}
}

func ensureHttpMethod(writer http.ResponseWriter, request *http.Request) bool {
	if request.Method != http.MethodPost {
		const msg = "only POST supported"
		http.Error(writer, msg, http.StatusMethodNotAllowed)
		glog.Warning(msg, ", request from=", request.RemoteAddr)
		return true
	}
	return false
}

func handleError(writer http.ResponseWriter, code int, description string, err error) {
	msg := fmt.Sprintf(description+": %v", err)
	glog.Error(msg)
	http.Error(writer, msg, code)
}