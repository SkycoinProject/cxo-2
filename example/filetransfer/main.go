package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/SkycoinPro/cxo-2-node/pkg/model"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/go-homedir"
	"github.com/skycoin/skycoin/src/cipher"
)

const (
	host               = "127.0.0.1"
	port               = 8080
	notifyRoute        = "/notify"
	cxoNodeRegisterUrl = "http://127.0.0.1:6420/api/v1/registerApp"
)

var storagePath string
var notifyRequest model.NotifyAppRequest

func main() {
	registerAppOnCxoNode()
	storagePath = initStoragePath()
	startServer()
}

func registerAppOnCxoNode() {
	client := http.DefaultClient
	appRequest := model.RegisterAppRequest{
		Address: fmt.Sprintf("%s:%v%s", host, port, notifyRoute),
		Name:    "File Transfer App",
	}

	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(appRequest)
	resp, err := client.Post(cxoNodeRegisterUrl, "application/json", b)
	if err != nil {
		fmt.Println("App registration on cxo node failed due to error: ", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusCreated {
		fmt.Printf("App registration on cxo node failed with status: %v. Make sure cxo node is running properly.", resp.StatusCode)
		os.Exit(1)
	}
	fmt.Println("App registered on cxo node successfully.")
}

func initStoragePath() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		processError(fmt.Errorf("unable to find working directory due to error: %v", err))
	}

	path := filepath.Join(homeDir, "cxo-file-transfer")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		errDir := os.MkdirAll(path, 0755)
		if errDir != nil {
			processError(fmt.Errorf("unable to prepare storage directory: %v due to error: %v", path, errDir))
		}
	}
	return path
}

func startServer() {
	router := gin.Default()
	router.POST("/notify", notify)

	if err := router.Run(fmt.Sprintf("%s:%v", host, port)); err != nil {
		panic(err.Error())
	}
}

func notify(c *gin.Context) {
	if err := c.BindJSON(&notifyRequest); err != nil {
		fmt.Println("Error receiving parcel due to error: ", err.Error())
		c.AbortWithStatusJSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	processData()
	fmt.Println("Local storage updated successfully...")
}

func processData() {
	publisherStoragePath := createStoragePathForPublisher(notifyRequest.RootHash.Publisher)
	cleanPublishersStorage(publisherStoragePath)
	storeDataOnPath(notifyRequest.RootHash.ObjectHeaderHash, publisherStoragePath)
}

func createStoragePathForPublisher(publisher string) string {
	publisherStoragePath := filepath.Join(storagePath, publisher)
	if _, err := os.Stat(publisherStoragePath); os.IsNotExist(err) {
		if errDir := os.Mkdir(publisherStoragePath, os.ModePerm); errDir != nil {
			processError(fmt.Errorf("unable to prepare storage directory: %v due to err: %v", publisherStoragePath, err))
		}
	}

	return publisherStoragePath
}

func cleanPublishersStorage(path string) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		processError(fmt.Errorf("reading storage path: %v failed due to error: %v", path, err))
	}
	for _, inf := range infos {
		if err := os.RemoveAll(filepath.Join(path, inf.Name())); err != nil {
			processError(fmt.Errorf("cleaning storage path: %v failed due to error: %v", path, err))
		}
	}
}

func storeDataOnPath(headerHash, path string) {
	header, err := retrieveHeaderByHash(headerHash)
	if err != nil {
		processError(err)
	}
	name := name(header)
	if isDirectory(header) {
		path = filepath.Join(path, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if errDir := os.Mkdir(path, os.ModePerm); errDir != nil {
				processError(errDir)
			}
		}

		for _, ref := range header.ExternalReferences {
			storeDataOnPath(ref, path)
		}
	} else {
		filePath := filepath.Join(path, name)
		object, err := retrieveObjectByHash(header.ObjectHash)
		if err != nil {
			processError(err)
		}
		createFile(filePath, object.Data)
	}
}

func retrieveHeaderByHash(hash string) (model.ObjectHeader, error) {
	for _, header := range notifyRequest.Parcel.ObjectHeaders {
		headerHash, err := sha256(header)
		if err != nil {
			return model.ObjectHeader{}, err
		}
		if hash == headerHash {
			return header, nil
		}
	}
	return model.ObjectHeader{}, fmt.Errorf("no object header found for hash: %v", hash)
}

func retrieveObjectByHash(hash string) (model.Object, error) {
	for _, object := range notifyRequest.Parcel.Objects {
		objectHash, err := sha256(object)
		if err != nil {
			return model.Object{}, err
		}
		if hash == objectHash {
			return object, nil
		}
	}
	return model.Object{}, fmt.Errorf("no corresponding object found for hash: %v", hash)
}

func sha256(object interface{}) (string, error) {
	b, err := json.Marshal(object)
	if err != nil {
		return "", err
	}

	sha256 := cipher.SumSHA256(b)
	return sha256.Hex(), nil
}

func name(oh model.ObjectHeader) string {
	for _, meta := range oh.Meta {
		if meta.Key == "name" {
			return meta.Value
		}
	}
	return ""
}

func isDirectory(oh model.ObjectHeader) bool {
	for _, meta := range oh.Meta {
		if meta.Key == "type" && meta.Value == "directory" {
			return true
		}
	}
	return false
}

func createFile(path string, content []byte) {
	f, err := os.Create(path)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()
	if _, err := f.Write(content); err != nil {
		panic(err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
}

func processError(err error) {
	fmt.Println(err)
	os.Exit(2)
}
