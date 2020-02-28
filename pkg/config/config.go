package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/SkycoinPro/cxo-2-node/pkg/util"
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
}

const (
	appRootFolderName   = ".cxo-node"
	keysFileName        = "keys.txt"
	configFileName      = "cxo-node-config.yml"
	defaultDiscoveryURL = "http://dmsg.discovery.skywire.cc" //"http://localhost:9090"
	defaultTrackerURL   = "dmsg://036cbf1297c2433303909674e1bc25ce341ec1c16012ba28a265066847960e2514:8084"
	serverPort          = uint16(8083)
)

// LoadConfig - load node's configuration
func LoadConfig(local *bool) Config {
	var appRootFolderPath = ""
	if *local == true {
		//retrieve working dir
		dir, err := os.Getwd()
		if err != nil {
			processError("unable to retrieve working directory from which node is running", err)
		}
		appRootFolderPath = filepath.Join(dir, appRootFolderName)
	} else {
		homeDir, err := homedir.Dir()
		if err != nil {
			processError("unable to find home directory", err)
		}
		appRootFolderPath = filepath.Join(homeDir, appRootFolderName)
	}

	//create root dir if not exist
	if _, err := os.Stat(appRootFolderPath); os.IsNotExist(err) {
		if errDir := os.Mkdir(appRootFolderPath, os.ModePerm); errDir != nil {
			processError("error creating cxo root directory.", errDir)
		}
	}

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
	}
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
