package main

import (
	cfg "hk4e/common/config"

	"github.com/spf13/cobra"

	"context"

	"hk4e/robot/app"
)

func RobotCmd() *cobra.Command {
	var configFile string
	app.APPVERSION = VERSION
	c := &cobra.Command{
		Use:   "robot",
		Short: "robot server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.InitConfig(configFile)
			return app.Run(context.Background())
		},
	}
	c.Flags().StringVar(&configFile, "config", "application.toml", "config file")
	return c
}
