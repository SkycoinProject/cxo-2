package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/model"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
	"github.com/SkycoinProject/dmsg/cipher"
	coincipher "github.com/SkycoinProject/skycoin/src/cipher"
)

// Service - node service model
type Service struct {
	config config.Config
	db     data
}

// NewService - initialize node service
func NewService(cfg config.Config) *Service {
	return &Service{
		config: cfg,
		db:     defaultData(),
	}
}

var notifyRoute = "/notify"

// Run - start's node service
func (s *Service) Run() {
	httpS := dmsghttp.Server{
		PubKey:    s.config.PubKey,
		SecKey:    s.config.SecKey,
		Port:      s.config.Port,
		Discovery: s.config.Discovery,
	}

	log.Infof("Starting cxo node with public key: %s and port: %v", s.config.PubKey.Hex(), s.config.Port)

	// prepare server route handling
	mux := http.NewServeMux()
	mux.HandleFunc(notifyRoute, s.notifyHandler)

	// run the server
	sErr := make(chan error, 1)
	sErr <- httpS.Serve(mux)
	close(sErr)
}

func (s *Service) notifyHandler(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["hash"]
	if !ok || len(keys[0]) < 1 {
		err := errors.New("missing root hash param")
		fmt.Println("error while receiving new hash from cxo tracker service with error: ", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	dataHash := keys[0]
	fmt.Println("Received new data hash from cxo tracker service: ", dataHash)

	go func() {
		time.Sleep(3 * time.Second)
		s.requestData(dataHash)
	}()

	w.WriteHeader(http.StatusOK)
}

func (s *Service) requestData(dataHash string) {
	if s.db.isSaved(dataHash) {
		fmt.Printf("received data object with hash: %v already exist", dataHash)
		return
	}
	sPK, sSK := cipher.GenerateKeyPair()
	client := dmsghttp.DMSGClient(s.config.Discovery, sPK, sSK)

	url := fmt.Sprint(s.config.TrackerAddress, "/data/request?hash=", dataHash)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("error creating requestData request for data hash: ", dataHash)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("requestData request failed due to error: ", err)
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading data: ", err)
		return
	}

	var parcel model.Parcel
	if dataObjectErr := json.Unmarshal(data, &parcel); dataObjectErr != nil {
		fmt.Println("error unmarshaling received data object with hash: ", dataHash)
		return
	}

	storagePath := s.config.StoragePath
	for _, header := range parcel.ObjectHeaders {
		name := dataHash
		isDirectory := false
		for _, meta := range header.Meta {
			if meta.Key == "type" && meta.Value == "directory" {
				isDirectory = true
			} else if meta.Key == "name" {
				name = fmt.Sprintf("%s_%s", dataHash[:8], meta.Value)
			}
		}
		if isDirectory {
			i := strings.Index(name, "_") + 1 // find split used when defining name and drop it if exists
			name = name[i:]
			storagePath = filepath.Join(storagePath, name)
			if _, err := os.Stat(storagePath); os.IsNotExist(err) {
				os.Mkdir(storagePath, os.ModePerm)
			}
		} else {
			filePath := filepath.Join(storagePath, name)
			//TODO replace with local data object storage
			object, err := findObjectByHash(parcel.Objects, header.ObjectHash)
			if err != nil {
				fmt.Printf("error writing file to local storage - can't find content for file %s", name)
			} else {
				createFile(filePath, object.Data)
			}
		}
	}

	// TODO store actual objects and headers for p2p communication
	fmt.Println("Storing file on file system finished successfully")
}

func findObjectByHash(objects []model.Object, objectHash string) (model.Object, error) {
	for _, object := range objects {
		bytes, err := json.Marshal(object)
		if err != nil {
			return model.Object{}, err
		}

		sha256 := coincipher.SumSHA256(bytes)
		if sha256.Hex() == objectHash {
			return object, nil
		}
	}
	return model.Object{}, fmt.Errorf("couldn't find object by hash %s", objectHash)
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
