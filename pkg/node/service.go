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

	"github.com/SkycoinPro/cxo-2-node/pkg/config"
	"github.com/SkycoinPro/cxo-2-node/pkg/model"
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
	objectHeaderHash := keys[0]
	fmt.Println("Received new object header hash from cxo tracker service: ", objectHeaderHash)

	go func() {
		time.Sleep(3 * time.Second)
		s.requestData(objectHeaderHash)
	}()

	w.WriteHeader(http.StatusOK)
}

func (s *Service) requestData(rootObjectHeaderHash string) {

	if s.db.isSaved(rootObjectHeaderHash) {
		fmt.Printf("received root object header hash: %v already exist", rootObjectHeaderHash)
		return
	}
	sPK, sSK := cipher.GenerateKeyPair()
	client := dmsghttp.DMSGClient(s.config.Discovery, sPK, sSK)

	if err := s.retrieveHeaders(client, rootObjectHeaderHash); err != nil {
		fmt.Printf("request data failed: %v", err)
	}

	//TODO populate this with retrieved headers
	var missingObjectHeaders []model.ObjectHeader

	storagePath := s.config.StoragePath
	for _, header := range missingObjectHeaders {
		name := rootObjectHeaderHash
		isDirectory := false
		for _, meta := range header.Meta {
			if meta.Key == "type" && meta.Value == "directory" {
				isDirectory = true
			} else if meta.Key == "name" {
				name = meta.Value
			}
		}
		if isDirectory {
			storagePath = filepath.Join(storagePath, name)
			if _, err := os.Stat(storagePath); os.IsNotExist(err) {
				_ = os.Mkdir(storagePath, os.ModePerm)
			}
		} else {
			filePath := filepath.Join(storagePath, name)
			object, err := s.fetchObject(client, header.ObjectHash)
			if err != nil {
				fmt.Printf("error writing file to local storage - can't fetch content for file %s", name)
			} else {
				err = s.db.saveObject(header.ObjectHash, filePath)
				if err != nil {
					fmt.Print("error saving object in db with hash: ", header.ObjectHash)
				}
				createFile(filePath, object.Data)
			}
		}
	}

	fmt.Println("Update of local storage finished successfully")
}

func (s *Service) retrieveHeaders(client *http.Client, headerHashes ...string) error {
	headers, err := s.fetchObjectHeaders(client, headerHashes...)
	if err != nil {
		return fmt.Errorf("fetching object headers with hashes: %v from service failed due to error: %v", headerHashes, err)
	}
	var missingHeaderHashes []string
	for i, header := range headers {
		for _, ref := range header.ExternalReferences {
			_, err := s.db.getObjectHeader(ref)
			if err != nil {
				if err == errCannotFindObjectHeader {
					missingHeaderHashes = append(missingHeaderHashes, ref)
					continue
				}
				return fmt.Errorf("fetching object header with hash: %v from db failed due to error: %v", ref, err)
			}
		}
		if err := s.db.saveObjectHeader(headerHashes[i], header); err != nil {
			return fmt.Errorf("saving object header with hash: %v failed due to error: %v", headerHashes[i], err)
		}
	}

	if len(missingHeaderHashes) > 0 {
		if err := s.retrieveHeaders(client, missingHeaderHashes...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) fetchObjectHeaders(client *http.Client, objectHeaderHashes ...string) ([]model.ObjectHeader, error) {
	objectHeadersResp := model.GetObjectHeadersResponse{}

	baseUrl := fmt.Sprint(s.config.TrackerAddress, "/data/object/header")
	params := ""
	for _, hash := range objectHeaderHashes {
		params = fmt.Sprint(params, "?hash=", hash)
	}

	url := fmt.Sprint(baseUrl, params)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []model.ObjectHeader{}, fmt.Errorf("error creating request for fetching object headers with hashes: %v", objectHeaderHashes)
	}

	resp, err := client.Do(req)
	if err != nil {
		return []model.ObjectHeader{}, fmt.Errorf("request for object headers with hashes: %v failed due to error: %v", objectHeaderHashes, err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []model.ObjectHeader{}, fmt.Errorf("error reading data: %v", err)
	}

	if objectHeaderErr := json.Unmarshal(data, &objectHeadersResp); objectHeaderErr != nil {
		return []model.ObjectHeader{}, fmt.Errorf("error unmarshaling received object headers response for with hashes: %v", objectHeaderHashes)
	}

	return objectHeadersResp.ObjectHeaders, nil
}

func (s *Service) fetchObject(client *http.Client, objectHash string) (model.Object, error) {
	object := model.Object{}
	url := fmt.Sprint(s.config.TrackerAddress, "/data/object?hash=", objectHash)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return object, fmt.Errorf("error creating request for object with hash: %v", objectHash)
	}

	resp, err := client.Do(req)
	if err != nil {
		return object, fmt.Errorf("request for object failed due to error: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			panic(err)
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return object, fmt.Errorf("error reading data: %v", err)
	}

	if objectErr := json.Unmarshal(data, &object); objectErr != nil {
		return object, fmt.Errorf("error unmarshaling received object with hash: %v", objectHash)
	}

	return object, nil
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
