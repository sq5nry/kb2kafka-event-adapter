package entity

type KBEvent struct {
	EventType  string `json:"eventType"`  //type of event (as defined by the ExtBusEventType enum)
	ObjectType string `json:"objectType"` //type of object being updated (as defined by the ObjectType enum)
	AccountId  string `json:"accountId"`  //account id being updated
	ObjectId   string `json:"objectId"`   //object id being updated
	MetaData   string `json:"metaData"`   //event-specific metadata, serialize as JSON
}
