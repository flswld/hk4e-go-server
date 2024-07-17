package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	cfg "hk4e/common/config"
	"hk4e/robot/app"
)

var (
	config = flag.String("config", "application.toml", "config file")
)

var VERSION = "UNKNOWN"

func main() {
	flag.Parse()
	app.APPVERSION = VERSION
	cfg.InitConfig(*config)
	err := app.Run(context.TODO())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
