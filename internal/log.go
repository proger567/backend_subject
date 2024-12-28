package internal

import (
	"context"
	"github.com/sirupsen/logrus"
)

type ErrorHandler interface {
	Handle(ctx context.Context, err error)
}

type UnitLogHandler struct {
	logger logrus.Logger
}

func NewUnitLogHandler(logger logrus.Logger) *UnitLogHandler {
	return &UnitLogHandler{
		logger: logger,
	}
}

func (uh UnitLogHandler) Handle(ctx context.Context, err error) {
	uh.logger.Log(logrus.ErrorLevel, err)
}
