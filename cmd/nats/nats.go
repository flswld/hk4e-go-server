package nats

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	cfg "hk4e/common/config"

	"github.com/nats-io/nats-server/v2/server"
)

func RunNatsServer(ctx context.Context) error {
	natsAddr := strings.ReplaceAll(cfg.GetConfig().MQ.NatsUrl, "nats://", "")
	if strings.Contains(natsAddr, ",") {
		return errors.New("not support nats cluster")
	}
	split := strings.Split(natsAddr, ":")
	if len(split) != 2 {
		return errors.New("nats addr format error")
	}
	host := split[0]
	port, err := strconv.Atoi(split[1])
	if err != nil {
		return err
	}

	opts := &server.Options{
		Host:                  host,
		Port:                  port,
		NoLog:                 false,
		NoSigs:                true,
		MaxControlLine:        4096,
		DisableShortFirstPing: true,
		Trace:                 false,
		Debug:                 true,
	}
	natsServer, err := server.NewServer(opts)
	if err != nil {
		return err
	}
	natsServer.ConfigureLogger()
	go natsServer.Start()
	ok := natsServer.ReadyForConnections(time.Second * 5)
	if !ok {
		return errors.New("nats server start error")
	}

	c := make(chan os.Signal, 1)
	if !cfg.GetConfig().Hk4e.StandaloneModeEnable {
		signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	}
	for {
		select {
		case <-ctx.Done():
			return nil
		case s := <-c:
			switch s {
			case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
				return nil
			case syscall.SIGHUP:
			default:
				return nil
			}
		}
	}
}
