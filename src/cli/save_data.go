package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"
	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/model"
	dmsgcipher "github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/skycoin/src/cipher"
	"github.com/spf13/cobra"
)

func publishDataCmd(client *client.TrackerClient, config config.Config) *cobra.Command {
	saveDataCmd := &cobra.Command{
		Short:                 "Publish data to the CXO Tracker service",
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
			publishDataRequest, err := prepareRequest(filePath, config)
			if err != nil {
				return err
			}
			err = client.SaveData(publishDataRequest)
			if err != nil {
				return err
			}

			fmt.Println("New data published successfully...")
			return nil
		},
	}

	return saveDataCmd
}

func prepareRequest(filePath string, config config.Config) (model.PublishDataRequest, error) {
	object, err := constructObject(filePath)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	fileName := filepath.Base(filePath)
	manifest, err := constructManifest(object, fileName)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	header, err := constructHeader(object, manifest)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	dataObject := model.DataObject{
		Header:   header,
		Manifest: manifest,
		Object:   object,
	}

	rootHash, err := constructRootHash(dataObject, config)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	return model.PublishDataRequest{
		RootHash:   rootHash,
		DataObject: dataObject,
	}, nil

}

func constructObject(filePath string) (model.Object, error) {
	object := model.Object{}
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return model.Object{}, fmt.Errorf("reading file: %v failed with error: %v", filePath, err)
	}

	object.Length = uint64(len(bytes))
	object.Data = bytes
	return object, nil
}

func constructManifest(object model.Object, fileName string) (model.Manifest, error) {
	objectStructure, err := constructObjectStructure(object)
	if err != nil {
		return model.Manifest{}, err
	}

	return model.Manifest{
		Length: 1, //FIXME - revisit this
		Hashes: []model.ObjectStructure{objectStructure},
		Meta:   []string{fileName},
	}, nil

}

func constructObjectStructure(object model.Object) (model.ObjectStructure, error) {
	objectHash, err := sha256(object)
	if err != nil {
		return model.ObjectStructure{}, fmt.Errorf("hashing object faled due to err: %v", err)
	}
	return model.ObjectStructure{
		Index:                   1, //FIXME - revisit this
		Hash:                    objectHash,
		Size:                    object.Length,
		RecursiveSizeFirstLevel: object.Length,
		RecursiveSizeFirstTotal: object.Length,
	}, nil
}

func constructHeader(object model.Object, manifest model.Manifest) (model.Header, error) {
	manifestHash, err := sha256(manifest)
	if err != nil {
		return model.Header{}, fmt.Errorf("hashing manifest faled due to err: %v", err)
	}

	return model.Header{
		Timestamp:    time.Now(),
		ManifestHash: manifestHash,
		ManifestSize: manifest.Length,
		DataHash:     manifest.Hashes[0].Hash, //FIXME - in future probably should be different
		DataSize:     object.Length,
	}, nil
}

func constructRootHash(dataObject model.DataObject, config config.Config) (model.RootHash, error) {
	signature, err := signDataObject(dataObject, config.SecKey)
	if err != nil {
		return model.RootHash{}, err
	}

	dataObjectHash, err := sha256(dataObject)
	if err != nil {
		return model.RootHash{}, fmt.Errorf("hashing data object faled due to err: %v", err)
	}

	return model.RootHash{
		Publisher:      config.PubKey.Hex(),
		Signature:      signature,
		Sequence:       1, //FIXME - sequence should be set in different way
		DataObjectHash: dataObjectHash,
	}, nil
}

func signDataObject(dataObject model.DataObject, secKey dmsgcipher.SecKey) (string, error) {
	dataObjectBytes, err := json.Marshal(dataObject)
	if err != nil {
		return "", fmt.Errorf("marshal data object faled due to err: %v", err)
	}
	signature, err := dmsgcipher.SignPayload(dataObjectBytes, secKey)
	if err != nil {
		return "", fmt.Errorf("signing data object faled due to err: %v", err)
	}

	return signature.Hex(), nil
}

func sha256(object interface{}) (string, error) {
	bytes, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	sha256 := cipher.SumSHA256(bytes) //FIXME - for some reason cannot use same method from dmsgchipher package
	return sha256.Hex(), nil
}
