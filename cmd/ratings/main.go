package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	kafkaConf "kb2kafka-event-adapter/config/kafka"
	listenerConf "kb2kafka-event-adapter/config/listener"
	"kb2kafka-event-adapter/domain/broker"
	"kb2kafka-event-adapter/domain/entity"
	"log" //TODO debug tracing instead of always evaluated printf
	"math/rand"
	"net/http"
	"os"
)

var kafkaConfig, kafkaConfigErr = kafkaConf.GetKafkaConfiguration()      //TODO not global, handle error
var listenerConfig, listenerConfigErr = listenerConf.GetListenerConfig() //TODO not global, handle error

var kafkaProducer = createProducer(kafkaConfig)

//TODO readiness & liveness probing
func main() {
	ensureConfigurationLoaded()
	http.HandleFunc(listenerConfig.EndpointPath, onHttpReq)

	log.Println("DEBUG: starting listener")
	if err := http.ListenAndServe(listenerConfig.ListenerAddress, nil); err != nil {
		log.Fatal("ERROR: " + err.Error())
	}
}

func ensureConfigurationLoaded() {
	if kafkaConfigErr != nil {
		log.Fatal("ERROR: " + kafkaConfigErr.Error())
	}
	if listenerConfigErr != nil {
		log.Fatal("ERROR: " + listenerConfigErr.Error())
	}
}

func onHttpReq(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		const msg = "only POST supported"
		http.Error(writer, msg, http.StatusMethodNotAllowed)
		log.Printf("WARN: %s, from=%s", msg, request.RemoteAddr)
	}

	var kbEvent entity.KBEvent
	err := json.NewDecoder(request.Body).Decode(&kbEvent)
	if err != nil {
		handleError(writer, http.StatusBadRequest, "decoding error", err)
		return
	}
	log.Printf("DEBUG: kbEvent from KB received: %#v\n", kbEvent)
	handleKbEvent(writer, request, kbEvent)
}

func handleKbEvent(writer http.ResponseWriter, request *http.Request, kbEvent entity.KBEvent) {
	ratingMsg := createRatingFromKbEvent(kbEvent)
	ratingMsgBytes, err := json.Marshal(ratingMsg)
	if err == nil {
		log.Printf("DEBUG: pushing KB request to kafka, id=%s, from=%s", ratingMsg.Id, request.RemoteAddr)
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
	ratingMsg.Body = event.MetaData;
	return ratingMsg
}

func handleError(writer http.ResponseWriter, code int, description string, err error) {
	var msg = fmt.Sprintf(description+": %v", err)
	log.Printf("ERROR: %s\n", msg)
	http.Error(writer, msg, code)
}

func dispatchToKafka(message []byte, key string, producer *kafka.Producer) error {
	deliveryChan := make(chan kafka.Event)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaConfig.ConsumerTopicName, Partition: kafka.PartitionAny},
		Value:          message,
		Key:            []byte(key),
	}, deliveryChan)
	if err != nil {
		log.Printf("ERROR: creating a message for delivery failed: %v\n", err)
		return err
	}

	event := <-deliveryChan
	ack := event.(*kafka.Message)
	if ack.TopicPartition.Error != nil {
		log.Printf("ERROR: delivery failed: %v\n", ack.TopicPartition.Error)
		return errors.New("delivery failed")
	} else {
		log.Printf("DEBUG: delivered message [%s] with key [%s] to topic %s [%d] at offset %v\n",
			message, key, *ack.TopicPartition.Topic, ack.TopicPartition.Partition, ack.TopicPartition.Offset)
		return nil
	}
}

func createProducer(config *kafkaConf.KafkaConfiguration) *kafka.Producer {
	host, err := os.Hostname()
	if err != nil {
		log.Fatal("ERROR: " + err.Error())
	}

	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServer,
		"client.id":         host,
	})
	if err != nil {
		log.Fatal(err)
	}
	return kafkaProducer
}
