package service

import (
	"testing"

	"go.uber.org/goleak"
	"route256/cart/internal/pkg/util/logger"
)

func TestMain(m *testing.M) {
	logger.New()
	goleak.VerifyTestMain(m)
}
