package cmd

import (
	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the server",
	Long:  "Stops the running server",
	RunE: func(cmd *cobra.Command, args []string) error {
		logs.Logger.Info(server.ID)
		if err := server.StopServer(); err != nil {
			return err
		}
		close(stopServer)
		return nil
	},
}
