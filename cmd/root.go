package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "dfs",
		Short: "A Distributed File Storage System",
		Long:  `A Distributed File Storage System, It can be deployed over a wide netowrk.`,
		Run: func(cmd *cobra.Command, args []string) {
			logs.Logger.Info("Welcome to world of distributed system")
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logs.Logger.Errorln(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(startCmd)
	rootCmd.AddCommand(storeCmd)
	rootCmd.AddCommand(stopCmd)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM)
	go func() {
		for sig := range c {
			if sig == syscall.SIGTERM {
				close(stopServer)
			}
		}
	}()
}
