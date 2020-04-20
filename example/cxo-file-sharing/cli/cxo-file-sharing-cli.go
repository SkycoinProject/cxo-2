package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/SkycoinProject/cxo-2/pkg/model"
	"github.com/SkycoinProject/skycoin/src/cipher"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func main() {
	cli := &cobra.Command{
		Short: fmt.Sprintf("The cxo-file-sharing command line interface"),
		Use:   fmt.Sprintf("cxo-file-sharing-cli"),
	}

	commands := []*cobra.Command{
		publishDataCmd(),
	}

	cli.Version = "0.1.1"
	cli.AddCommand(commands...)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}

func publishDataCmd() *cobra.Command {
	publishDataCmd := &cobra.Command{
		Short:                 "Publish files",
		Use:                   "publish [flags] [path_to_file]",
		Long:                  "Publish new data trough CXO NODE",
		SilenceUsage:          true,
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		RunE: func(c *cobra.Command, args []string) error {
			filePath := args[0]
			if filePath == "" {
				return c.Help()
			}

			publishDataRequest, err := prepareRequest(filePath)
			if err != nil {
				return err
			}

			b, err := json.MarshalIndent(publishDataRequest, "", "  ")
			if err != nil {
				return err
			}

			f, err := ioutil.TempFile(os.TempDir(), "publishDataRequest.json")
			if err != nil {
				return err
			}

			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}

				if err := os.Remove(f.Name()); err != nil {
					panic(err)
				}
			}()

			if _, err := f.Write(b); err != nil {
				panic(err)
			}

			if err = f.Sync(); err != nil {
				panic(err)
			}

			tempFilePath, err := filepath.Abs(f.Name())
			fmt.Println(tempFilePath)
			if err != nil {
				fmt.Println("failed to retrieve temp file name due to error", err)
				return err
			}

			cmd := exec.Command("cxo-node-cli", "publish", tempFilePath)
			var cxoNodeOut, cxoNodeErrOut bytes.Buffer
			cmd.Stdout = &cxoNodeOut
			cmd.Stderr = &cxoNodeErrOut
			if err := cmd.Run(); err != nil {
				fmt.Println("cxo node publish failed due to error", err)
				return err
			}

			fmt.Println(cxoNodeErrOut.String())
			fmt.Println(cxoNodeOut.String())
			return nil
		},
	}

	return publishDataCmd
}

func prepareRequest(filePath string) (model.PublishDataRequest, error) {
	parcel := model.Parcel{}

	// we're supporting only one path in the request at a time
	processPath(&parcel, filePath, []int{})

	rootHash, err := constructRootHash(parcel)
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

func constructRootHash(parcel model.Parcel) (model.RootHash, error) {
	firstObjectHeaderHash, err := sha256(parcel.ObjectHeaders[0])
	if err != nil {
		return model.RootHash{}, fmt.Errorf("hashing first object header faled due to err: %v", err)
	}

	return model.RootHash{
		Timestamp:        time.Now(),
		ObjectHeaderHash: firstObjectHeaderHash,
	}, nil
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
