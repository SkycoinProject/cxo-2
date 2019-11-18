package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/SkycoinPro/cxo-2-node/src/cli/client"
	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/model"
	dmsgcipher "github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/skycoin/src/cipher"
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

			seqNo, err := client.GetNewSequenceNumber(config.PubKey.Hex())
			if err != nil {
				fmt.Println("Could not find latest sequence number due to error ", err)
				return err
			}

			publishDataRequest, err := prepareRequest(filePath, config, seqNo)
			if err != nil {
				return err
			}

			b, err := json.MarshalIndent(publishDataRequest, "", "  ")
			if err != nil {
				return err
			}
			_, err = os.Stdout.Write(b)
			if err != nil {
				return err
			}

			err = client.PublishData(publishDataRequest)
			if err != nil {
				return err
			}

			fmt.Println("New data published successfully...")
			return nil
		},
	}

	return publishDataCmd
}

func prepareRequest(filePath string, config config.Config, sequenceNumber uint64) (model.PublishDataRequest, error) {
	parcel := model.Parcel{}

	filePaths := strings.Split(filePath, ",")
	for _, path := range filePaths {
		isDir, err := isDirectory(path)
		if err != nil {
			fmt.Printf("Unable to parse path %s due to error %v", path, err)
			continue
		}
		if isDir {
			processDirectory(&parcel, path)
		} else {
			processFile(&parcel, path)
		}
	}

	rootHash, err := constructRootHash(parcel, config, sequenceNumber)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	return model.PublishDataRequest{
		RootHash: rootHash,
		Parcel:   parcel,
	}, nil

}

func processFile(parcel *model.Parcel, path string) {
	object, err := constructObject(path)
	if err != nil {
		processError("Error constructing object", err)
	}

	fileName := filepath.Base(path)
	objectHeader, err := constructObjectHeader(object, fileName)
	if err != nil {
		processError("Error constructing object header", err)
	}

	parcel.ObjectHeaders = append(parcel.ObjectHeaders, objectHeader)
	parcel.Objects = append(parcel.Objects, object)
}

func constructObject(path string) (model.Object, error) {
	object := model.Object{}
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return model.Object{}, fmt.Errorf("reading file: %v failed with error: %v", path, err)
	}

	object.Length = uint64(len(bytes))
	object.Data = bytes
	return object, nil
}

func constructObjectHeader(object model.Object, fileName string) (model.ObjectHeader, error) {
	objectHeader := model.ObjectHeader{}
	objectHash, err := sha256(object)
	if err != nil {
		return model.ObjectHeader{}, fmt.Errorf("hashing object faled due to err: %v", err)
	}

	metaStructure := model.Meta{
		Key:   "type",
		Value: "file",
	}

	metaDescription := model.Meta{
		Key:   "name",
		Value: fileName,
	}

	objectHeader.ObjectHash = objectHash
	objectHeader.ObjectSize = object.Length
	objectHeader.Meta = []model.Meta{metaStructure, metaDescription}

	return objectHeader, nil
}

func constructRootHash(parcel model.Parcel, config config.Config, sequenceNumber uint64) (model.RootHash, error) {
	signature, err := signParcel(parcel, config.PubKey, config.SecKey)
	if err != nil {
		return model.RootHash{}, err
	}

	firstObjectHeaderHash, err := sha256(parcel.ObjectHeaders[0])
	if err != nil {
		return model.RootHash{}, fmt.Errorf("hashing first object header faled due to err: %v", err)
	}

	return model.RootHash{
		Publisher:        config.PubKey.Hex(),
		Signature:        signature,
		Timestamp:        time.Now(),
		Sequence:         sequenceNumber,
		ObjectHeaderHash: firstObjectHeaderHash,
	}, nil
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

//func constructManifest(objects []model.Object, fileName string) (model.Manifest, error) {
//	objectsStructures, err := constructObjectsStructures(objects)
//	if err != nil {
//		return model.Manifest{}, err
//	}
//
//	meta := []string{fileName} //FIXME - currently only one file name is taken but should be changed in future
//	length := len(objectsStructures) + len(meta)
//
//	return model.ExternalReferences{
//		Length: uint64(length),
//		Hashes: objectsStructures,
//		Meta:   meta,
//	}, nil
//
//}

//func constructObjectsStructures(objects []model.Object) ([]model.ObjectStructure, error) {
//	var structures []model.ObjectStructure
//	for i, object := range objects {
//		objectHash, err := sha256(object)
//		if err != nil {
//			return []model.ObjectStructure{}, fmt.Errorf("hashing object faled due to err: %v", err)
//		}
//		structures = append(structures, model.ObjectStructure{
//			Index:                   uint64(i),
//			Hash:                    objectHash,
//			Size:                    object.Length,
//			RecursiveSizeFirstLevel: object.Length,
//			RecursiveSizeFirstTotal: object.Length,
//		})
//	}
//	return structures, nil
//}

//func constructHeader(object model.DataObject, manifest model.ExternalReferences) (model.Header, error) {
//	manifestHash, err := sha256(manifest)
//	if err != nil {
//		return model.Header{}, fmt.Errorf("hashing manifest faled due to err: %v", err)
//	}
//
//	return model.Header{
//		ManifestHash: manifestHash,
//		ManifestSize: manifest.Length,
//		DataHash:     manifest.Hashes[0].Hash, //FIXME - in future probably should be different
//		DataSize:     object.Length,
//	}, nil
//}

func sha256(object interface{}) (string, error) {
	bytes, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	sha256 := cipher.SumSHA256(bytes) //FIXME - for some reason cannot use same method from dmsgchipher package
	return sha256.Hex(), nil
}

func isDirectory(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		fmt.Println(err)
		return false, err
	}

	switch mode := fi.Mode(); {
	case mode.IsDir():
		// do directory stuff
		return true, nil
	case mode.IsRegular():
		// do file stuff
		return false, nil
	default:
		return false, fmt.Errorf("couldn't determe type of object due to err: %v", err)
	}
}

func processDirectory(parcel *model.Parcel, path string) {

}

func processError(message string, err error) {
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(2)
}
