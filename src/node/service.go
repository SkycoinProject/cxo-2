package node

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/model"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
	"github.com/SkycoinProject/dmsg/cipher"
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

	data, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error reading data: ", err)
		return
	}

	var dataObject model.Parcel
	if dataObjectErr := json.Unmarshal(data, &dataObject); dataObjectErr != nil {
		fmt.Println("error unmarshaling received data object with hash: ", dataHash)
		return
	}
	fileName := dataHash
	// FIXME add directory support here
	actualNameSet := false
	for _, header := range dataObject.ObjectHeaders {
		for _, meta := range header.Meta {
			if meta.Key == "name" {
				fileName = fmt.Sprintf("%s_%s", dataHash[:8], meta.Value)
				actualNameSet = true
				break
			}
		}
		if actualNameSet {
			break
		}
	}

	filePath := filepath.Join(s.config.StoragePath, fileName)
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	if _, err := f.Write(dataObject.Objects[0].Data); err != nil {
		panic(err)
	}
	if err = f.Sync(); err != nil {
		panic(err)
	}
	if err := s.db.saveObject(dataHash, s.config.StoragePath); err != nil {
		fmt.Println("error saving data object in db with hash: ", dataHash)
		return
	}
	fmt.Println("Storing file on file system finished successfully")
}
