package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	"kb2kafka-event-adapter/config"
	"kb2kafka-event-adapter/domain/broker"
	"kb2kafka-event-adapter/domain/confirmation"
	"kb2kafka-event-adapter/domain/entity"
	kafkaRepo "kb2kafka-event-adapter/domain/repository"
	"kb2kafka-event-adapter/infrastructure"
	"net/http"
	"time"
)

const HEALTHCHECK_PATH = "/health"

var configuration *config.Configuration
var kafkaProducer *kafka.Producer
var kafkaConsumer *kafka.Consumer

func init() {
	flag.Parse() //workaround for glog "logging before flag.Parse" issue
	configuration = config.GetConfiguration()
	kafkaProducer = kafkaRepo.NewProducer(configuration.KafkaConfiguration)
	kafkaConsumer = confirmation.NewConfirmationConsumer(configuration.KafkaConfiguration)
}

type Users struct {
	configuration *config.Configuration
}

func main() {
	http.HandleFunc(HEALTHCHECK_PATH, func(w http.ResponseWriter, r *http.Request) { infrastructure.OnHealthCheck(w, configuration.ListenerConfiguration) })
	http.HandleFunc(configuration.ListenerConfiguration.EndpointPath, onHttpReq)

	glog.Info("starting listener")
	if err := http.ListenAndServe(configuration.ListenerConfiguration.ListenerAddress, nil); err != nil {
		glog.Exit(err.Error())
	}
}

/*
HTTP request handler. Main program loop.
*/
func onHttpReq(writer http.ResponseWriter, request *http.Request) {
	if ensureHttpMethod(writer, request) {
		return
	}

	kbEvent, err := extractKbEvent(request)
	if err != nil {
		sendErrorResponse(writer, http.StatusBadRequest, "decoding error", err)
		return
	}

	ratingMsg, err := storeKbEvent(kbEvent)
	if err != nil {
		sendErrorResponse(writer, http.StatusInternalServerError, "could not dispatch rating message", err)
		return
	}

	if err := handleConfirmationMessage(ratingMsg); err != nil {
		sendErrorResponse(writer, http.StatusInternalServerError, "could not confirm persistence result", err)
	}
}

/*
1st step of the processing loop. Acquire KB notification event.
*/
func extractKbEvent(request *http.Request) (*broker.KBEvent, error) {
	var kbEvent broker.KBEvent
	err := json.NewDecoder(request.Body).Decode(&kbEvent)
	glog.Infof("KbEvent received [%#v]", kbEvent)
	return &kbEvent, err
}

/*
2nd step of the processing loop. Store KB event data into kafka.
*/
func storeKbEvent(kbEvent *broker.KBEvent) (*entity.RatingMessage, error) {
	ratingMsg := entity.CreateRatingFromKbEvent(kbEvent)
	ratingMsgBytes, err := json.Marshal(ratingMsg)
	if err == nil {
		glog.Infof("pushing KB request to kafka, id=%s", ratingMsg.Id)
		err = kafkaRepo.Dispatch(ratingMsgBytes, ratingMsg.Id, kafkaProducer, configuration.KafkaConfiguration)
		if err != nil {
			return ratingMsg, err
		}
	} else {
		return &entity.RatingMessage{}, err
	}
	return ratingMsg, nil
}

/*
3rd step of the processing loop. Await confirmation from the persistence.
*/
func handleConfirmationMessage(ratingMsg *entity.RatingMessage) error {
	cfmMessage, err := confirmation.GetMessage(kafkaConsumer, 2*time.Second) //TODO upfront assumption for local testing only
	if err == nil {
		if !confirmation.ValidateConfirmationMessage(cfmMessage, ratingMsg) {
			return err
		} else {
			return nil
		}
	} else {
		return err
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

func sendErrorResponse(writer http.ResponseWriter, code int, description string, err error) {
	msg := fmt.Sprintf(description+": %v", err)
	glog.Error(msg)
	http.Error(writer, msg, code)
}
