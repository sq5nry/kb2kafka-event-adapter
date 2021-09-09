package dependency

import (
	"kb2kafka-event-adapter/domain/repository"
)

func NewRatingRepository(db interface{}) repository.RatingRepository {
	//switch connection := db.(type) {
	//case *pg.DB:
	//	return postgres.NewRatingRepository(connection)
	//}

	return nil
}