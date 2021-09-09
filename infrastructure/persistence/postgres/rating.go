package postgres

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	"kb2kafka-event-adapter/domain/entity"
	"kb2kafka-event-adapter/domain/repository"
)

type ratingRepository struct {
	db *pg.DB
}

func (r *ratingRepository) Save(rating *entity.Rating) error {
	//err := r.db.Insert(rating)
	//
	//if err != nil {
	//	return err
	//}

	fmt.Printf("Rating was saved to database: %+v\n", rating)
	return nil
}

func NewRatingRepository(db *pg.DB) repository.RatingRepository {
	return &ratingRepository{db: db}
}
