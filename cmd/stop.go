package cmd

import (
	"os"
	"strconv"
	"syscall"

	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the server",
	Long:  "Stops the running server",
	RunE: func(cmd *cobra.Command, args []string) error {
		pidFile := "/tmp/dfs_server.pid"
		pidData, err := os.ReadFile(pidFile)
		if err != nil {
			logs.Logger.Fatalf("Failed to read PID file: %v", err)
		}
		pid, err := strconv.Atoi(string(pidData))
		if err != nil {
			logs.Logger.Fatalf("Invalid PID in file: %v", err)
		}

		// Send a stop signal to the server process
		proc, err := os.FindProcess(pid)
		if err != nil {
			logs.Logger.Fatalf("Failed to find process: %v", err)
		}

		err = proc.Signal(syscall.SIGTERM)
		if err != nil {
			logs.Logger.Fatalf("Failed to send stop signal: %v", err)
		}
		logs.Logger.Info("Gracefully Shutting Down the Server")
		return nil
	},
}
