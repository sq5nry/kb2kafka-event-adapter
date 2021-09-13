package dependency

//func NewRatingConsumer(configuration kafka2.KafkaConfiguration) broker.RatingConsumer  {
//
//	kafkaConsumer, err := kafka.NewConsumer(&kafka.ConfigMap{
//		"bootstrap.servers" : configuration.BootstrapServer,
//		"group.id": configuration.GroupId,
//		"auto.offset.reset": "earliest",
//	})
//
//
//	if err != nil {
//		panic(err)
//	}
//
//	kafkaConsumer.Subscribe(configuration.ConsumerTopicName, nil)
//
//	return kafka_broker.NewRatingConsumer(kafkaConsumer)
//}
