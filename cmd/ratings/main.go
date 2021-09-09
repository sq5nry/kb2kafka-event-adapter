package main

import (
	"encoding/json"
	"fmt"
	"github.com/Shopify/sarama"
	"kb2kafka-event-adapter/config"
	//"kb2kafka-event-adapter/dependency"
	//"kb2kafka-event-adapter/domain/interactor"
	"log"
	"net/http"
)

func main() {
	//kafkaProperties := config.ReadKafkaConfigurationFromFile()
	listenerConfig := config.GetListenerConfig()

	http.HandleFunc(listenerConfig.EndpointPath, onPost)

	log.Println("DEBUG: starting listener")
	if err := http.ListenAndServe(listenerConfig.ListenerAddress, nil); err != nil {
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
		//eventBytes, err := json.Marshal(event)
		log.Printf("DEBUG: event from KB received: %#v\n", event)

		var ratingMsg RatingMessage;
		ratingMsg.Id = event.AccountId	//TODO account or object id?
		ratingMsg.RecipeId = event.AccountId	//string(eventBytes)	//TODO should be in "value", putting it here for now
		ratingMsg.Value = 123

		ratingMsgBytes, err := json.Marshal(ratingMsg)
		if err == nil {
			//TODO async
			//TODO topic from config
			log.Printf("DEBUG: pushing KB request to kafka, id=%s, from=%s", ratingMsg.Id, request.RemoteAddr)
			err := pushKbEventToQueue("test-topic", ratingMsgBytes, ratingMsg.Id)
			if err != nil {
				var msg = fmt.Sprintf("could not push to kafka: %v", err)
				log.Println("ERROR: " + msg)
				http.Error(writer, msg, http.StatusInternalServerError)
			}
		} else {
			var msg = fmt.Sprintf("could not create a rating message: %v", err)
			log.Println("ERROR: " + msg)
			http.Error(writer, msg, http.StatusInternalServerError)
		}

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

func pushKbEventToQueue(topic string, message []byte, key string) error {
	brokerUrl := []string{"localhost:9092"}	//TODO config
	producer, err := connectProducer(brokerUrl)
	if err != nil {
		return err
	}
	defer producer.Close()
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
		Key: sarama.StringEncoder(key),
	}

	partition, offset, err := producer.SendMessage(msg)
	if err != nil {
		return err
	}
	log.Printf("DEBUG: Message is stored in topic(%s)/partition(%d)/offset(%d)\n", topic, partition, offset)
	return nil
}
