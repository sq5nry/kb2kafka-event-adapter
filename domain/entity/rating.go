package entity

import (
	"github.com/golang/glog"
	"github.com/google/uuid"
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
	if len(event.AccountId) == 0 {
		event.AccountId = uuid.New().String()
		glog.Infof("KB event has no id, assigning id=%s", event.AccountId)
	}

	return &RatingMessage{
		Id:       event.AccountId,	//TODO assuming KB.AccountId
		RecipeId: event.AccountId,	//TODO assuming KB.AccountId
		Value:    int8(rand.Int()),	//TODO field needed?
		Body:     event.MetaData,
	}
}