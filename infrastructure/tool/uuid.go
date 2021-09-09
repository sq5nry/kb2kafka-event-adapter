package tool

import (
	"github.com/google/uuid"
	"kb2kafka-event-adapter/domain/tool"
)

type uuidGenerator struct {
}

func (g *uuidGenerator) Generate() string {
	return uuid.New().String()
}

func NewUuidGenerator() tool.IdGenerator {
	return &uuidGenerator{}
}