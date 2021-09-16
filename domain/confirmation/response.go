package confirmation

import (
	"encoding/json"
)

type ConfirmationMessage struct {
	Id      string `json:"id"`
	Message string `json:"message"`		//TODO align with Piotrek's part
}

func CreateConfirmationMessageFromJSON(data []byte) (*ConfirmationMessage, error) {
	var cfmMessage *ConfirmationMessage
	err := json.Unmarshal(data, &cfmMessage)
	if err != nil {
		return &ConfirmationMessage{}, err
	} else {
		return cfmMessage, nil
	}
}