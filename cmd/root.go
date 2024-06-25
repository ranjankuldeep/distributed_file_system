package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/spf13/cobra"
)

var (
	rootPersistFlag bool
	rootLocalFlag   bool
	rootCmd         = &cobra.Command{
		Use:   "dfs",
		Short: "A Distributed File Storage System",
		Long:  `A Distributed File Storage System, It can be deployed over a wide netowrk.`,
		Run: func(cmd *cobra.Command, args []string) {
			logs.Logger.Info("Hello ")
		},
	}
	echo = &cobra.Command{
		Use:   "echo everything",
		Short: "a naive echo example",
		Long:  "a naive long echo",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Echo:" + strings.Join(args, " "))
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
	rootCmd.PersistentFlags().BoolVarP(&rootPersistFlag, "persistFlag", "p", false, "a peristent root flag")
	rootCmd.LocalFlags().BoolVarP(&rootLocalFlag, "localFlag", "l", false, "a local root flag")
	rootCmd.AddCommand(echo)
}
