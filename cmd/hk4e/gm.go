package main

import (
	"context"

	cfg "hk4e/common/config"
	"hk4e/gm/app"

	"github.com/spf13/cobra"
)

func GMCmd() *cobra.Command {
	var configFile string
	c := &cobra.Command{
		Use:   "gm",
		Short: "gm server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.InitConfig(configFile)
			return app.Run(context.Background())
		},
	}
	c.Flags().StringVar(&configFile, "config", "application.toml", "config file")
	return c
}
