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

func announceDataCmd(client *client.TrackerClient, config config.Config) *cobra.Command {
	announceDataCmd := &cobra.Command{
		Short:                 "Announce new data to the CXO Tracker service",
		Use:                   "announce [flags] [path_to_file]",
		Long:                  "Announce new data to the CXO Tracker service",
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			filePath := args[0]
			if filePath == "" {
				return c.Help()
			}
			announceDataRequest, err := prepareRequest(filePath, config)
			if err != nil {
				return err
			}
			err = client.AnnounceData(announceDataRequest)
			if err != nil {
				return err
			}

			fmt.Println("New data announced successfully...")
			return nil
		},
	}

	return announceDataCmd
}

func prepareRequest(filePath string, config config.Config) (model.AnnounceDataRequest, error) {
	object, err := constructObject(filePath)
	if err != nil {
		return model.AnnounceDataRequest{}, err
	}

	fileName := filepath.Base(filePath)
	manifest, err := constructManifest([]model.Object{object}, fileName)
	if err != nil {
		return model.AnnounceDataRequest{}, err
	}

	header, err := constructHeader(object, manifest)
	if err != nil {
		return model.AnnounceDataRequest{}, err
	}

	dataObject := model.DataObject{
		Header:   header,
		Manifest: manifest,
		Objects:  []model.Object{object},
	}

	rootHash, err := constructRootHash(dataObject, config)
	if err != nil {
		return model.AnnounceDataRequest{}, err
	}

	return model.AnnounceDataRequest{
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

func constructManifest(objects []model.Object, fileName string) (model.Manifest, error) {
	objectsStructures, err := constructObjectsStructures(objects)
	if err != nil {
		return model.Manifest{}, err
	}

	meta := []string{fileName} //FIXME - currently only one file name is taken but should be changed in future
	length := len(objectsStructures) + len(meta)

	return model.Manifest{
		Length: uint64(length),
		Hashes: objectsStructures,
		Meta:   meta,
	}, nil

}

func constructObjectsStructures(objects []model.Object) ([]model.ObjectStructure, error) {
	var structures []model.ObjectStructure
	for i, object := range objects {
		objectHash, err := sha256(object)
		if err != nil {
			return []model.ObjectStructure{}, fmt.Errorf("hashing object faled due to err: %v", err)
		}
		structures = append(structures, model.ObjectStructure{
			Index:                   uint64(i),
			Hash:                    objectHash,
			Size:                    object.Length,
			RecursiveSizeFirstLevel: object.Length,
			RecursiveSizeFirstTotal: object.Length,
		})
	}
	return structures, nil
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

	return model.RootHash{
		Publisher: config.PubKey.Hex(),
		Signature: signature,
		Sequence:  1, //FIXME - sequence should be set in different way
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
