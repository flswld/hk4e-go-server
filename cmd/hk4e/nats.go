package main

import (
	"context"

	"hk4e/cmd/nats"
	cfg "hk4e/common/config"

	"github.com/spf13/cobra"
)

func NatsCmd() *cobra.Command {
	var configFile string
	c := &cobra.Command{
		Use:   "nats",
		Short: "nats server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg.InitConfig(configFile)
			return nats.RunNatsServer(context.Background())
		},
	}
	c.Flags().StringVar(&configFile, "config", "application.toml", "config file")
	return c
}
