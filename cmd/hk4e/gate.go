package main

import (
	"context"

	cfg "hk4e/common/config"
	"hk4e/gate/app"

	"github.com/spf13/cobra"
)

func GateCmd() *cobra.Command {
	var configFile string
	app.APPVERSION = VERSION
	c := &cobra.Command{
		Use:   "gate",
		Short: "gate server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.InitConfig(configFile)
			return app.Run(context.Background())
		},
	}
	c.Flags().StringVar(&configFile, "config", "application.toml", "config file")
	return c
}
