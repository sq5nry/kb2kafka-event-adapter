package broker

type RatingMessage struct {
	Id       string `json:"id"`
	RecipeId string `json:"recipe_id"`
	Value    int8   `json:"value"`
	Body     string `json:"body"`	//TODO either change type for Value and remove this or use this as KB data container
}
