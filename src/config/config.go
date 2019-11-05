package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SkycoinPro/cxo-2-node/src/util"
	"github.com/mitchellh/go-homedir"
	"github.com/skycoin/dmsg/cipher"
	"github.com/skycoin/dmsg/disc"
)

// Config - node's configuration model
type Config struct {
	TrackerAddress string
	PubKey         cipher.PubKey
	SecKey         cipher.SecKey
	Port           uint16
	Discovery      disc.APIClient
	StoragePath    string
}

const (
	appRootFolderName   = ".cxo-node"
	storageFolderName   = "files"
	keysFileName        = "keys.txt"
	defaultDiscoveryURL = "http://dmsg.discovery.skywire.skycoin.com"
	serverPort          = uint16(8083)
)

var appRootFolderPath string

// LoadConfig - load node's configuration
func LoadConfig() Config {
	storagePath := initializeAppFolderStructure()

	keysFilePath := filepath.Join(appRootFolderPath, keysFileName)
	sPK, sSK := util.PrepareKeyPair(keysFilePath)

	return Config{
		TrackerAddress: getTrackerAddress(),
		PubKey:         sPK,
		SecKey:         sSK,
		Port:           serverPort,
		Discovery:      disc.NewHTTP(defaultDiscoveryURL),
		StoragePath:    storagePath,
	}
}

func initializeAppFolderStructure() string {
	homeDir, err := homedir.Dir()
	if err != nil {
		panic(err) //FIXME - maybe returning error and print to console would be better
	}

	appRootFolderPath = filepath.Join(homeDir, appRootFolderName)
	storagePath := filepath.Join(appRootFolderPath, storageFolderName)
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		errDir := os.MkdirAll(storagePath, 0755)
		if errDir != nil {
			panic(err)
		}
	}

	return storagePath
}

// FIXME cxo tracker address is hardcoded for now until we find better solution
func getTrackerAddress() string {
	trackerPubKey := "036cbf1297c2433303909674e1bc25ce341ec1c16012ba28a265066847960e2514"
	trackerPort := "8084"
	return fmt.Sprintf("dmsg://%v:%v", trackerPubKey, trackerPort)
}
