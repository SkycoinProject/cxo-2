package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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
	sPK, sSK := prepareKeys()

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

func prepareKeys() (cipher.PubKey, cipher.SecKey) {
	var sPK cipher.PubKey
	var sSK cipher.SecKey

	keysFilePath := filepath.Join(appRootFolderPath, keysFileName)
	data, err := ioutil.ReadFile(keysFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File containing keys doesn't exist. Generating new keys pairs...")
		} else {
			fmt.Printf("Reading keys from system failed due to error: %v. Generating new keys pairs...", err)
		}
		sPK, sSK = cipher.GenerateKeyPair()
		storeKeys(sPK, sSK, keysFilePath)
		return sPK, sSK
	}

	keys := strings.Split(string(data), ":")
	_ = sPK.UnmarshalText([]byte(keys[0]))
	_ = sSK.UnmarshalText([]byte(keys[1]))

	return sPK, sSK
}

func storeKeys(sPK cipher.PubKey, sSK cipher.SecKey, keysFilePath string) {
	fmt.Println("Trying to store keys on file system")
	f, err := os.Create(keysFilePath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := f.WriteString(fmt.Sprintf("%v:%v", sPK.Hex(), sSK.Hex())); err != nil {
		panic(err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
	fmt.Println("Storing keys on file system finished successfully")
}

// FIXME cxo tracker address is hardcoded for now until we find better solution
func getTrackerAddress() string {
	trackerPubKey := "02150fc16da944e94cf15d79600790e717c2cf106d7e80ba601e0bdf6438a89b83" // pub changes every time cx tracker is restarted
	trackerPort := "8084"
	return fmt.Sprintf("dmsg://%v:%v", trackerPubKey, trackerPort)
}
