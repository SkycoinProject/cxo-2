package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SkycoinPro/cxo-2-node/pkg/cli/client"
	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/model"
	dmsgcipher "github.com/SkycoinProject/dmsg/cipher"
	"github.com/spf13/cobra"
)

func publishDataCmd(client *client.TrackerClient, config config.Config) *cobra.Command {
	publishDataCmd := &cobra.Command{
		Short:                 "Publish new data to the CXO Tracker service",
		Use:                   "publish [flags] [path_to_file]",
		Long:                  "Publish new data to the CXO Tracker service",
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			filePath := args[0]
			if filePath == "" {
				return c.Help()
			}

			var req model.PublishDataRequest

			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				fmt.Println("Could not read request file ", err)
				return err
			}

			if err := json.Unmarshal(b, &req); err != nil {
				fmt.Println("Could not read unmarshal file ", err)
				return err
			}

			seqNo, err := client.GetNewSequenceNumber(config.PubKey.Hex())
			if err != nil {
				fmt.Println("Could not find latest sequence number due to error ", err)
				return err
			}

			sig, err := signParcel(req.Parcel, config.PubKey, config.SecKey)
			if err != nil {
				fmt.Println("Signing parcel failed due to error ", err)
				return err
			}

			req.RootHash.Publisher = config.PubKey.Hex()
			req.RootHash.Sequence = seqNo
			req.RootHash.Signature = sig

			reqBytes, err := json.MarshalIndent(req, "", "  ")
			if err != nil {
				return err
			}

			_, err = os.Stdout.Write(reqBytes)
			if err != nil {
				return err
			}

			err = client.PublishData(req)
			if err != nil {
				return err
			}

			return nil
		},
	}

	return publishDataCmd
}

func signParcel(parcel model.Parcel, pubKey dmsgcipher.PubKey, secKey dmsgcipher.SecKey) (string, error) {
	parcelBytes, err := json.Marshal(parcel)
	if err != nil {
		return "", fmt.Errorf("marshal parcel faled due to err: %v", err)
	}
	signature, err := dmsgcipher.SignPayload(parcelBytes, secKey)
	if err != nil {
		return "", fmt.Errorf("signing parcel faled due to err: %v", err)
	}

	if err = dmsgcipher.VerifyPubKeySignedPayload(pubKey, signature, parcelBytes); err != nil {
		return "", fmt.Errorf("parcel signature verification failed due to error: %v", err)
	}

	return signature.Hex(), nil
}
