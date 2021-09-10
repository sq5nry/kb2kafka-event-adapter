package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"kb2kafka-event-adapter/config"
	"log"
	"net/http"
)

type KBEvent struct {
	EventType  string `json:"eventType"`  //type of event (as defined by the ExtBusEventType enum)
	ObjectType string `json:"objectType"` //type of object being updated (as defined by the ObjectType enum)
	AccountId  string `json:"accountId"`  //account id being updated
	ObjectId   string `json:"objectId"`   //object id being updated
	MetaData   string `json:"metaData"`   //event-specific metadata, serialize as JSON
}

type RatingMessage struct {
	Id       string `json:"id"`
	RecipeId string `json:"recipe_id"`
	Value    int8   `json:"value"`
}

var kafkaConfig = config.ReadKafkaConfigurationFromFile()
var kafkaProducer kafka.Producer = createProducer(&kafkaConfig)

func createProducer(config *config.KafkaConfiguration) kafka.Producer {
	kafkaProducer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": config.BootstrapServer,
	})
	if err != nil {
		log.Fatal(err)
	}

	return *kafkaProducer
}

func main() {
	listenerConfig := config.GetListenerConfig()

	http.HandleFunc(listenerConfig.EndpointPath, onPost)

	log.Println("DEBUG: starting listener")
	if err := http.ListenAndServe(listenerConfig.ListenerAddress, nil); err != nil {
		log.Fatal(err)
	}
}

func onPost(writer http.ResponseWriter, request *http.Request) {
	switch request.Method {
	case "POST":
		var kbEvent KBEvent
		err := json.NewDecoder(request.Body).Decode(&kbEvent)
		if err != nil {
			handleError(writer, http.StatusBadRequest, "decoding error", err)
			return
		}

		log.Printf("DEBUG: kbEvent from KB received: %#v\n", kbEvent)
		ratingMsg := createRatingFromKbEvent(kbEvent)
		ratingMsgBytes, err := json.Marshal(ratingMsg)
		if err == nil {
			log.Printf("DEBUG: pushing KB request to kafka, id=%s, from=%s", ratingMsg.Id, request.RemoteAddr)
			err = dispatchToKafka(ratingMsgBytes, ratingMsg.Id, &kafkaProducer)
			if err != nil {
				handleError(writer, http.StatusInternalServerError, "could not push to kafka", err)
			}
		} else {
			handleError(writer, http.StatusInternalServerError, "could not create a rating message", err)
		}

	default:
		const msg = "only POST supported"
		http.Error(writer, msg, http.StatusMethodNotAllowed)
		log.Printf("WARN: %s, from=%s", msg, request.RemoteAddr)
	}
}

func createRatingFromKbEvent(event KBEvent) RatingMessage {
	var ratingMsg RatingMessage
	ratingMsg.Id = event.AccountId       //TODO assuming KB.AccountId
	ratingMsg.RecipeId = event.AccountId //TODO assuming KB.AccountId
	ratingMsg.Value = 123                //TODO KB data should go here
	return ratingMsg
}

func handleError(writer http.ResponseWriter, code int, description string, err error) {
	var msg = fmt.Sprintf(description+": %v", err)
	log.Println("ERROR: " + msg)
	http.Error(writer, msg, code)
}

func dispatchToKafka(message []byte, key string, producer *kafka.Producer) error {
	deliveryChan := make(chan kafka.Event)
	err := producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &kafkaConfig.ConsumerTopicName, Partition: kafka.PartitionAny},
		Value:          message,
		Key: 			[]byte(key),
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
