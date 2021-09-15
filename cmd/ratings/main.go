package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/golang/glog"
	kafkaCfg "kb2kafka-event-adapter/config/kafka"
	listenerCfg "kb2kafka-event-adapter/config/listener"
	"kb2kafka-event-adapter/domain/broker"
	"kb2kafka-event-adapter/domain/entity"
	"math/rand"
	"net/http"
	"os"
)

var kafkaConfig *kafkaCfg.KafkaConfiguration
var listenerConfig *listenerCfg.ListenerConfiguration
var kafkaProducer *kafka.Producer

func init() {
	initializeGlog()

	var err error
	kafkaConfig, err = kafkaCfg.GetKafkaConfiguration()
	if err != nil {
		glog.Fatalf("error in kafka configuration: %#v", err)
	}

	listenerConfig, err = listenerCfg.GetListenerConfig()
	if err != nil {
		glog.Fatalf("error in listener configuration: %#v", err)
	}

	kafkaProducer = createProducer(kafkaConfig)
}

func initializeGlog() {
	flag.Parse()	//workaround for glog "logging before flag.Parse" message
}

//TODO readiness & liveness probing
func main() {
	http.HandleFunc(listenerConfig.EndpointPath, onHttpReq)

	glog.Info("starting listener")
	if err := http.ListenAndServe(listenerConfig.ListenerAddress, nil); err != nil {
		glog.Exit(err.Error())
	}
}

func onHttpReq(writer http.ResponseWriter, request *http.Request) {
	if ensureHttpMethod(writer, request) {
		return
	}

	kbEvent, done := extractKbData(writer, request)
	if done {
		return
	}

	handleKbEvent(writer, kbEvent)
}

func extractKbData(writer http.ResponseWriter, request *http.Request) (entity.KBEvent, bool) {
	var kbEvent entity.KBEvent
	err := json.NewDecoder(request.Body).Decode(&kbEvent)
	if err != nil {
		handleError(writer, http.StatusBadRequest, "decoding error", err)
		return entity.KBEvent{}, true
	}
	glog.Info("KbEvent received", kbEvent)
	return kbEvent, false
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

func handleKbEvent(writer http.ResponseWriter, kbEvent entity.KBEvent) {
	ratingMsg := createRatingFromKbEvent(kbEvent)
	ratingMsgBytes, err := json.Marshal(ratingMsg)
	if err == nil {
		glog.Infof("pushing KB request to kafka, id=%s", ratingMsg.Id)
		err = dispatchToKafka(ratingMsgBytes, ratingMsg.Id, kafkaProducer)
		if err != nil {
			handleError(writer, http.StatusInternalServerError, "could not push to kafka", err)
		}
	} else {
		handleError(writer, http.StatusInternalServerError, "could not create a rating message", err)
	}
}

func createRatingFromKbEvent(event entity.KBEvent) broker.RatingMessage {
	var ratingMsg broker.RatingMessage
	ratingMsg.Id = event.AccountId       //TODO assuming KB.AccountId
	ratingMsg.RecipeId = event.AccountId //TODO assuming KB.AccountId
	ratingMsg.Value = int8(rand.Int())   //TODO field needed?
	ratingMsg.Body = event.MetaData
	return ratingMsg
}

func handleError(writer http.ResponseWriter, code int, description string, err error) {
	glog.Error(err.Error())
	http.Error(writer, fmt.Sprintf(description+": %v", err), code)
}

func dispatchToKafka(message []byte, key string, producer *kafka.Producer) error {
	deliveryChan := make(chan kafka.Event)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaConfig.ConsumerTopicName, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            []byte(key),
	}, deliveryChan)
	if err != nil {
		glog.Warningf("creating a message for delivery failed: %#v\n", err)
		return err
	}

	event := <-deliveryChan
	ack := event.(*kafka.Message)
	if ack.TopicPartition.Error != nil {
		glog.Warning("delivery failed: ", ack.TopicPartition.Error)
		return errors.New("delivery failed")
	} else {
		glog.Infof("delivered message [%s] with key [%s] to topic %s [%d] at offset %v\n",
			message, key, *ack.TopicPartition.Topic, ack.TopicPartition.Partition, ack.TopicPartition.Offset)
		return nil
	}
}

func createProducer(config *kafkaCfg.KafkaConfiguration) *kafka.Producer {
	host, err := os.Hostname()
	if err != nil {
		glog.Fatal(err.Error())
	}

	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServer,
		"client.id":         host,
	})
	if err != nil {
		glog.Fatal(err)
	}
	return kafkaProducer
}
