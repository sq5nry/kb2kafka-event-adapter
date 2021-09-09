package dependency

import (
	"kb2kafka-event-adapter/domain/tool"
	tool2 "kb2kafka-event-adapter/infrastructure/tool"
)

func NewIdGenerator() tool.IdGenerator {
	return tool2.NewUuidGenerator()
}