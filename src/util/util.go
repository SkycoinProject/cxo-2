package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/skycoin/dmsg/cipher"
)

/* PrepareKeyPair - reads pub and sec key from specified file.
If file doesn't exists or reading keys failed new keys are generated and stored on file system */
func PrepareKeyPair(keysFilePath string) (cipher.PubKey, cipher.SecKey) {
	var sPK cipher.PubKey
	var sSK cipher.SecKey
	data, err := ioutil.ReadFile(keysFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Info("File containing keys doesn't exist. Generating new keys pairs...")
		} else {
			log.Infof("Reading keys from system failed due to error: %v. Generating new keys pairs...", err)
		}
		return generateKeyPairAndStore(keysFilePath)
	}

	keys := strings.Split(string(data), ":")
	pubKeyErr := sPK.UnmarshalText([]byte(keys[0]))
	secKeyErr := sSK.UnmarshalText([]byte(keys[1]))

	if pubKeyErr != nil || secKeyErr != nil {
		return generateKeyPairAndStore(keysFilePath)
	}
	return sPK, sSK
}

func generateKeyPairAndStore(keysFilePath string) (cipher.PubKey, cipher.SecKey) {
	sPK, sSK := cipher.GenerateKeyPair()
	log.Info("Trying to store keys on file system")
	f, err := os.Create(keysFilePath)
	if err != nil {
		log.Fatal("Creating file for storing key pair failed due to error: ", err)
	}

	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := f.WriteString(fmt.Sprintf("%v:%v", sPK.Hex(), sSK.Hex())); err != nil {
		log.Fatal("Writing keys to the file failed due to error: ", err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
	log.Info("Storing keys on file system finished successfully")
	return sPK, sSK
}
