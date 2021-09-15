package entity

import (
	"kb2kafka-event-adapter/domain/broker"
	"math/rand"
)

type RatingMessage struct {
	Id       string `json:"id"`
	RecipeId string `json:"recipe_id"`
	Value    int8   `json:"value"`
	Body     string `json:"body"`	//TODO either change type for Value and remove this or use this as KB data container
}

func CreateRatingFromKbEvent(event *broker.KBEvent) *RatingMessage {
	return &RatingMessage{
		Id:       event.AccountId,	//TODO assuming KB.AccountId
		RecipeId: event.AccountId,	//TODO assuming KB.AccountId
		Value:    int8(rand.Int()),	//TODO field needed?
		Body:     event.MetaData,
	}
}