package cli

import (
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"

	"github.com/spf13/cobra"
)

func saveDataCmd(client *client.TrackerClient) *cobra.Command {
	saveDataCmd := &cobra.Command{
		Short:                 "Save data to the CXO Tracker service",
		Use:                   "save [flags] [path_to_file]",
		Long:                  "Save new data to the CXO Tracker service",
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			filePath := args[0]
			if filePath == "" {
				return c.Help()
			}

			err := client.SaveData(filePath)

			switch err.(type) {
			case nil:
				fmt.Println("success")
				return nil
			default:
				return err
			}
		},
	}

	return saveDataCmd
}
