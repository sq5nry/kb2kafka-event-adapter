package repository

import "kb2kafka-event-adapter/domain/entity"

type RatingRepository interface {
	Save(*entity.Rating) error
}