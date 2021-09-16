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
	"net/http"
	"time"
)

var configuration *config.Configuration
var kafkaProducer *kafka.Producer
var kafkaConsumer *kafka.Consumer

func init() {
	flag.Parse() //workaround for glog "logging before flag.Parse" issue
	configuration = config.GetConfiguration()
	kafkaProducer = kafkaRepo.NewProducer(configuration.KafkaConfiguration)
	kafkaConsumer = confirmation.NewConfirmationConsumer(configuration.KafkaConfiguration)
}

//TODO health check probing
func main() {
	http.HandleFunc(configuration.ListenerConfiguration.EndpointPath, onHttpReq)

	glog.Info("starting listener")
	if err := http.ListenAndServe(configuration.ListenerConfiguration.ListenerAddress, nil); err != nil {
		glog.Exit(err.Error())
	}
}

func onHttpReq(writer http.ResponseWriter, request *http.Request) {
	if ensureHttpMethod(writer, request) {
		return
	}

	if kbEvent, err := extractKbEvent(request); err != nil {
		handleError(writer, http.StatusBadRequest, "decoding error", err)
	} else {
		glog.Info("KbEvent received", kbEvent)
		ratingMsg := storeKbEvent(writer, kbEvent)

		handleConfirmationMessage(writer, ratingMsg)
	}
}

/*
1st step of the processing loop. Acquire KB notification event.
 */
func extractKbEvent(request *http.Request) (*broker.KBEvent, error) {
	var kbEvent broker.KBEvent
	err := json.NewDecoder(request.Body).Decode(&kbEvent)
	return &kbEvent, err
}

/*
2nd step of the processing loop. Store KB event data into kafka.
*/
func storeKbEvent(writer http.ResponseWriter, kbEvent *broker.KBEvent) *entity.RatingMessage {
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
	return ratingMsg
}

/*
3rd step of the processing loop. Await confirmation from the persistence.
*/
func handleConfirmationMessage(writer http.ResponseWriter, ratingMsg *entity.RatingMessage) {
	cfmMessage, err := confirmation.GetMessage(kafkaConsumer, 2 * time.Second)	//TODO upfront assumption for local testing only
	if err != nil {
		handleError(writer, http.StatusInternalServerError, "confirmation message error", err)
	} else {
		if !confirmation.ValidateConfirmationMessage(cfmMessage, ratingMsg) {
			handleError(writer, http.StatusInternalServerError, "confirmation message error", err)
		}
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
