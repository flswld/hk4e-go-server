package main

import (
	"context"
	"flag"
	"fmt"
	_ "net/http/pprof"
	"os"

	cfg "hk4e/common/config"
	"hk4e/dispatch/app"
	"hk4e/pkg/statsviz_serve"
)

var (
	config = flag.String("config", "application.toml", "config file")
)

var VERSION = "UNKNOWN"

func main() {
	flag.Parse()
	go func() {
		_ = statsviz_serve.Serve("0.0.0.0:2345")
	}()
	app.APPVERSION = VERSION
	cfg.InitConfig(*config)
	err := app.Run(context.TODO())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
