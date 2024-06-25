package cmd

import (
	"log"

	"github.com/ranjankuldeep/distributed_file_system/logs"
	"github.com/ranjankuldeep/distributed_file_system/util"
	"github.com/spf13/cobra"
)

var (
	filePath string
)
var (
	storeCmd = &cobra.Command{
		Use:   "store",
		Short: "Store data in distributed file Storage",
		Long:  "Store ",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			<-serverReady
			filePath = args[0]
			key, err := util.GetFileName(filePath)
			if err != nil {
				logs.Logger.Errorf("Error reading file stat from the specified path %s", filePath)
				log.Fatal(err)
			}

			file, err := util.ReadFile(filePath)
			if err != nil {
				logs.Logger.Errorf("Error reading file from the specified path %s", filePath)
				file.Close()
				log.Fatal(err)

			}
			if err := server.Store(key, file); err != nil {
				logs.Logger.Errorf("Error Storing file %+v", err)
				return err
			}
			logs.Logger.Info("File Stored Succesfully")
			file.Close()
			return nil
		},
	}
)
