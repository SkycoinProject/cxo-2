package cli

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/SkycoinPro/cxo-2-node/pkg/cli/client"
	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/model"
	dmsgcipher "github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/skycoin/src/cipher"
	"github.com/spf13/cobra"
)

func publishDataCmd(client *client.TrackerClient, config config.Config) *cobra.Command {
	publishDataCmd := &cobra.Command{
		Short:        "Publish new data to the CXO Tracker service",
		Use:          "publish [flags] [path_to_file]",
		Long:         "Publish new data to the CXO Tracker service",
		SilenceUsage: true,
		Args:         cobra.MinimumNArgs(1),
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

	// we're supporting only one path in the request at a time
	processPath(&parcel, filePath, []int{})

	rootHash, err := constructRootHash(parcel, config, sequenceNumber)
	if err != nil {
		return model.PublishDataRequest{}, err
	}

	return model.PublishDataRequest{
		RootHash: rootHash,
		Parcel:   parcel,
	}, nil

}

func processPath(parcel *model.Parcel, path string, parentDirectories []int) {
	isDir, err := isDirectory(path)
	if err != nil {
		fmt.Printf("Unable to parse path %s due to error %v", path, err)
		return
	}
	if isDir {
		headerIndex, subPaths := processDirectory(parcel, path, parentDirectories)
		for _, subPath := range subPaths {
			newParentDirectories := []int{}
			newParentDirectories = append(newParentDirectories, parentDirectories...)
			newParentDirectories = append(newParentDirectories, headerIndex)
			processPath(parcel, subPath, newParentDirectories)
		}
		if len(parentDirectories) > 0 {
			hash, err := sha256(parcel.ObjectHeaders[headerIndex])
			if err != nil {
				fmt.Printf("can't construct header hash for directory %s due to error %v", path, err)
			} else {
				constructExternalReference(parcel, hash, uint64(0), parentDirectories)
			}
		}
	} else {
		processFile(parcel, path, parentDirectories)
	}
}

func processDirectory(parcel *model.Parcel, path string, parentDirectories []int) (int, []string) {
	dirName := filepath.Base(path)
	dirIndex := len(parcel.ObjectHeaders)

	objectHeader, err := constructDirHeader(dirName)
	if err != nil {
		processError("Error constructing object header", err)
	}

	parcel.ObjectHeaders = append(parcel.ObjectHeaders, objectHeader)

	paths, err := listDirectory(path)
	if err != nil {
		processError("Not able to list directory "+path, err)
	}

	return dirIndex, paths //FIXME returning directory header index in headers list here (if not try to fall back to header hash or depth)
}

func processFile(parcel *model.Parcel, path string, parentDirectories []int) {
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

	if len(parentDirectories) == 0 {
		// if it's not in directory, just simple file, finish here
		return
	}

	hash, err := sha256(objectHeader)
	if err != nil {
		fmt.Printf("can't construct header hash for file %s due to error %v", path, err)
		return
	}

	constructExternalReference(parcel, hash, objectHeader.ObjectSize, parentDirectories)
}

func constructExternalReference(parcel *model.Parcel, hash string, size uint64, parentDirectories []int) {
	if len(parentDirectories) == 0 {
		return
	}

	parentIndex := parentDirectories[len(parentDirectories)-1]
	parentDirectoryHeader := parcel.ObjectHeaders[parentIndex]
	parentDirectoryHeader.ExternalReferences = append(parentDirectoryHeader.ExternalReferences, hash)
	parentDirectoryHeader.ExternalReferencesSize++
	parentDirectoryHeader.RecursiveSizeTotal += size
	parentDirectoryHeader.RecursiveSizeFirstLevel += size
	parcel.ObjectHeaders[parentIndex] = parentDirectoryHeader

	if size == 0 {
		// no need to update senior parents if the size is 0
		return
	}

	for i := len(parentDirectories) - 2; i >= 0; i-- {
		upTheTreeIndex := parentDirectories[i]
		parentDirectoryHeader := parcel.ObjectHeaders[upTheTreeIndex]
		parentDirectoryHeader.RecursiveSizeTotal += size
		parcel.ObjectHeaders[upTheTreeIndex] = parentDirectoryHeader
	}
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

func constructDirHeader(dirName string) (model.ObjectHeader, error) {
	objectHeader := model.ObjectHeader{}

	metaStructure := model.Meta{
		Key:   "type",
		Value: "directory",
	}
	metaDescription := model.Meta{
		Key:   "name",
		Value: dirName,
	}

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
		return true, nil
	case mode.IsRegular():
		return false, nil
	default:
		return false, fmt.Errorf("couldn't determe type of object due to err: %v", err)
	}
}

// TODO consider droping this func and use FileInfo instead of string paths
func listDirectory(path string) ([]string, error) {
	var files []string
	infos, err := ioutil.ReadDir(path)
	for _, inf := range infos {
		files = append(files, filepath.Join(path, inf.Name()))
	}
	return files, err
}

func processError(message string, err error) {
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(2)
}
