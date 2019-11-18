package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SkycoinPro/cxo-2-node/src/util"
	"github.com/SkycoinProject/dmsg/cipher"
	"github.com/SkycoinProject/dmsg/disc"
	"github.com/kelseyhightower/envconfig"
	"github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
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
	configFileName      = "cxo-node-config.yml"
	defaultDiscoveryURL = "http://dmsg.discovery.skywire.skycoin.com"
	defaultTrackerURL   = "dmsg://036cbf1297c2433303909674e1bc25ce341ec1c16012ba28a265066847960e2514:8084"
	serverPort          = uint16(8083)
)

// LoadConfig - load node's configuration
func LoadConfig() Config {
	homeDir, err := homedir.Dir()
	if err != nil {
		processError("unable to find home directory", err)
	}
	appRootFolderPath := filepath.Join(homeDir, appRootFolderName)

	storagePath := initStoragePath(appRootFolderPath)

	keysFilePath := filepath.Join(appRootFolderPath, keysFileName)
	sPK, sSK := util.PrepareKeyPair(keysFilePath)

	configFilePath := filepath.Join(appRootFolderPath, configFileName)
	confFile := configFile{
		TrackerURL:   defaultTrackerURL,
		DiscoveryURL: defaultDiscoveryURL,
	}
	readConfigFile(configFilePath, &confFile)
	readEnv(&confFile)
	//TODO consider validation of TrackerURL and DiscoveryURL

	return Config{
		TrackerAddress: confFile.TrackerURL,
		PubKey:         sPK,
		SecKey:         sSK,
		Port:           serverPort,
		Discovery:      disc.NewHTTP(confFile.DiscoveryURL),
		StoragePath:    storagePath,
	}
}

func initStoragePath(appPath string) string {
	storagePath := filepath.Join(appPath, storageFolderName)
	if _, err := os.Stat(storagePath); os.IsNotExist(err) {
		errDir := os.MkdirAll(storagePath, 0755)
		if errDir != nil {
			processError("unable to prepare storage directory", err)
		}
	}
	return storagePath
}

func readConfigFile(path string, conf *configFile) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("File containing config doesn't exist. Reading config from env variables...")
			return
		}
		processError("unable to open config file", err)
	}

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&conf)
	if err != nil {
		processError("unable to parse config file", err)
	}
}

func readEnv(cfg *configFile) {
	err := envconfig.Process("cxo_node", cfg)
	if err != nil {
		processError("unable to read environment variables", err)
	}
}

func processError(message string, err error) {
	fmt.Println(message)
	fmt.Println(err)
	os.Exit(2)
}

type configFile struct {
	TrackerURL   string `envconfig:"TRACKER_URL" yaml:"trackerUrl"`
	DiscoveryURL string `envconfig:"DISCOVERY_URL" yaml:"discoveryUrl"`
}
