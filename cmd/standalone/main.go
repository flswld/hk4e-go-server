package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hk4e/cmd/nats"
	cfg "hk4e/common/config"
	dispatchapp "hk4e/dispatch/app"
	gateapp "hk4e/gate/app"
	gmapp "hk4e/gm/app"
	gsapp "hk4e/gs/app"
	multiapp "hk4e/multi/app"
	nodeapp "hk4e/node/app"
	"hk4e/pkg/statsviz_serve"

	"github.com/flswld/halo/logger"
)

var (
	config = flag.String("config", "application.toml", "config file")
)

func main() {
	flag.Parse()
	go func() {
		_ = statsviz_serve.Serve("0.0.0.0:4567")
	}()
	cfg.InitConfig(*config)

	logger.InitLogger(&logger.Config{
		AppName:      "standalone",
		Level:        logger.ParseLevel(cfg.GetConfig().Logger.Level),
		TrackLine:    cfg.GetConfig().Logger.TrackLine,
		TrackThread:  cfg.GetConfig().Logger.TrackThread,
		EnableFile:   cfg.GetConfig().Logger.EnableFile,
		DisableColor: cfg.GetConfig().Logger.DisableColor,
		EnableJson:   cfg.GetConfig().Logger.EnableJson,
	})
	logger.Warn("standalone start")
	defer func() {
		logger.Warn("standalone exit")
		logger.CloseLogger()
	}()

	stopChan := make(chan struct{})

	ctxNats, cancelNats := context.WithCancel(context.Background())
	ctxNode, cancelNode := context.WithCancel(context.Background())
	ctxDispatch, cancelDispatch := context.WithCancel(context.Background())
	ctxGate, cancelGate := context.WithCancel(context.Background())
	ctxGs, cancelGs := context.WithCancel(context.Background())
	ctxGm, cancelGm := context.WithCancel(context.Background())
	ctxMulti, cancelMulti := context.WithCancel(context.Background())

	go func() {
		err := nats.RunNatsServer(ctxNats)
		if err != nil {
			panic(err)
		}
		stopChan <- struct{}{}
	}()

	time.Sleep(time.Second)

	go func() {
		err := nodeapp.Run(ctxNode)
		if err != nil {
			panic(err)
		}
		cancelNats()
	}()

	time.Sleep(time.Second)

	go func() {
		err := dispatchapp.Run(ctxDispatch)
		if err != nil {
			panic(err)
		}
		cancelNode()
	}()

	time.Sleep(time.Second)

	go func() {
		err := gateapp.Run(ctxGate)
		if err != nil {
			panic(err)
		}
		cancelDispatch()
	}()

	time.Sleep(time.Second)

	go func() {
		err := gsapp.Run(ctxGs)
		if err != nil {
			panic(err)
		}
		cancelGate()
	}()

	time.Sleep(time.Second)

	go func() {
		err := gmapp.Run(ctxGm)
		if err != nil {
			panic(err)
		}
		cancelGs()
	}()

	go func() {
		err := multiapp.Run(ctxMulti)
		if err != nil {
			panic(err)
		}
		cancelGm()
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		select {
		case s := <-c:
			logger.Warn("get a signal %s", s.String())
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				cancelMulti()
				<-stopChan
				return
			case syscall.SIGHUP:
			default:
				return
			}
		}
	}
}
