package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	//"kb2kafka-event-adapter/dependency"
	//"kb2kafka-event-adapter/domain/interactor"
	"log"
	"net/http"
)

func main() {
	//kafkaProperties := config.ReadKafkaConfigurationFromFile()

	http.HandleFunc("/endpoint", onPost)

	fmt.Printf("Starting server for testing HTTP POST...\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
	//db, err := dependency.NewPostgresConnection()
	//
	//if err != nil {
	//	fmt.Printf("%s\n", err.Error())
	//	return
	//} else {
	//	fmt.Println("Database connection was initialized.")
	//	defer dependency.Close(db)
	//}

	//ratingInteractor := interactor.NewRatingInteractor(
	//	dependency.NewRatingRepository(db),
	//	dependency.NewRatingConsumer(kafkaProperties),
	//	dependency.NewIdGenerator(),
	//)

	//fmt.Println("Start receiving from Kafka")
	//for {
	//	err := ratingInteractor.GetIncomingMessage()
	//
	//	if err != nil {
	//		fmt.Printf("error from ratingInteractor: [%s]\n", err)
	//	}
	//}
}

type KBEvent struct {
	Id       string `json:"id"`
	RecipeId string `json:"recipe_id"`
	Value    int8   `json:"value"`
	//SubscriptionId string `json:"subscription_id"`
	//Ekey           string `json:"ekey"`
	//Status         string `json:"status"`
	//EffectiveDate  string `json:"effective_date"`
}

func onPost(writer http.ResponseWriter, request *http.Request) {
	//TODO single condition
	switch request.Method {
	case "POST":
		var event KBEvent
		err := json.NewDecoder(request.Body).Decode(&event)
		if err != nil {
			//TODO provide a message
			http.Error(writer, err.Error(), http.StatusBadRequest)
			return
		}
		//TODO handle error
		eventBytes, err := json.Marshal(event)
		log.Printf("DEBUG: event from KB received: %#v\n", event)

		//TODO async
		//TODO topic from config
		pushKbEventToQueue("test-topic", eventBytes)
		log.Printf("DEBUG: kb request pushed to queue, from=%s", request.RemoteAddr)
	default:
		const msg = "only POST supported"
		http.Error(writer, msg, http.StatusMethodNotAllowed)
		log.Printf("WARN: %s, from=%s", msg, request.RemoteAddr)
	}
}

func connectProducer(brokersUrl []string) (sarama.SyncProducer,error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	// NewSyncProducer creates a new SyncProducer using the given broker addresses and configuration.
	conn, err := sarama.NewSyncProducer(brokersUrl, config)
	if err != nil {
		return nil, err
	}
	return conn, nil
}


func pushKbEventToQueue(topic string, message []byte) error {
	brokerUrl := []string{"localhost:9092"}	//TODO config
	producer, err := connectProducer(brokerUrl)
	if err != nil {
		return err
	}
	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}
