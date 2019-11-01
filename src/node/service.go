package node

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/skycoin/dmsg/cipher"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
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

	fmt.Println("Starting server with public key: ", s.config.PubKey.Hex())
	fmt.Println("Starting server with port: ", s.config.Port)

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

	var dataObject model.DataObject
	if dataObjectErr := json.Unmarshal(data, &dataObject); dataObjectErr != nil {
		fmt.Println("error unmarshaling received data object with hash: ", dataHash)
		return
	}
	fileName := dataHash
	if len(dataObject.Manifest.Meta) > 0 {
		fileName = fmt.Sprintf("%s_%s", dataHash[:8], dataObject.Manifest.Meta[0])
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

	if _, err := f.Write(data); err != nil {
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
