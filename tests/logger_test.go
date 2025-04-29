package tests

import (
	"testing"

	"github.com/flswld/halo/logger"
)

func TestLogger(t *testing.T) {
	logger.InitLogger(nil)
	defer logger.CloseLogger()
	logger.Warn("logger test ...")
	for i := 0; i < 10000; i++ {
		logger.Info("%v", i)
	}
}
