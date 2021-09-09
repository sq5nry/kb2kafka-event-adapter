package postgres

import (
	"github.com/stretchr/testify/suite"

	"kb2kafka-event-adapter/domain/tool"
	tool3 "kb2kafka-event-adapter/infrastructure/tool"
	"testing"
)

type RatingRepositoryTestSuite struct {
	suite.Suite
	//db          *pg.DB
	idGenerator tool.IdGenerator
}

func (suite *RatingRepositoryTestSuite) SetupTest() {
	//host, port, user, password, database := config.GetDatabaseConfig()

	//suite.db = pg.Connect(&pg.Options{
	//	Addr:     fmt.Sprintf("%s:%s", host, port),
	//	User:     user,
	//	Password: password,
	//	Database: database,
	//})
	suite.idGenerator = tool3.NewUuidGenerator()
}

func (suite *RatingRepositoryTestSuite) TestSave() {
	//rating := &entity.Rating{
	//	Id:       suite.idGenerator.Generate(),
	//	RecipeId: suite.idGenerator.Generate(),
	//	Value:    5,
	//}
	//
	//underTest := ratingRepository{db: suite.db}
	//
	//err := underTest.Save(rating)
	//
	//assert.Nil(suite.T(), err)
}

func TestRatingRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RatingRepositoryTestSuite))
}
