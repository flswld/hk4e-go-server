package tests

import (
	"hk4e/pkg/logger"

	"testing"
)

func TestLogger(t *testing.T) {
	logger.InitLogger(nil)
	defer logger.CloseLogger()
	logger.Warn("logger test ...")
	for i := 0; i < 10000; i++ {
		logger.Info("%v", i)
	}
}
