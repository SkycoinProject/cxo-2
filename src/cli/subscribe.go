package cli

import (
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"
	"github.com/spf13/cobra"
)

func subscribeCmd(client *client.TrackerClient) *cobra.Command {
	subscribeCmd := &cobra.Command{
		Short:                 "Subscribe to public key",
		Use:                   "subscribe [flags] [public_key]",
		Long:                  "Subscribe to public key on CXO Tracker service",
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			pubKey := args[0]
			if pubKey == "" {
				return c.Help()
			}

			err := client.Subscribe(pubKey)

			switch err.(type) {
			case nil:
				fmt.Println("success")
				return nil
			default:
				return err
			}
		},
	}

	return subscribeCmd
}
