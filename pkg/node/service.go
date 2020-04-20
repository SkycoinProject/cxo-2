package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/SkycoinProject/cxo-2/pkg/errors"

	"github.com/SkycoinProject/cxo-2/pkg/config"
	"github.com/SkycoinProject/cxo-2/pkg/model"
	"github.com/SkycoinProject/cxo-2/pkg/node/data"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
	dmsgcipher "github.com/SkycoinProject/dmsg/cipher"
	log "github.com/sirupsen/logrus"
)

// Service - node service model
type Service struct {
	config config.Config
	db     data.Data
}

// NewService - initialize node service
func NewService(cfg config.Config) *Service {
	return &Service{
		config: cfg,
		db:     data.DefaultData(),
	}
}

var notifyRoute = "/notify"

// Run - start's node service
func (s *Service) Run() {
	webServer := InitServerAndController(s.db)
	go func() {
		webServer.Run()
	}()

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
	if r.Method != http.MethodPost {
		http.NotFound(w, r)
		return
	}
	var rootHash model.RootHash
	err := json.NewDecoder(r.Body).Decode(&rootHash)
	if err != nil {
		log.Error("Error while receiving new root hash: ", err)
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	fmt.Println("Received new root hash from cxo tracker service: ", rootHash.Key())

	go func() {
		time.Sleep(3 * time.Second)
		s.requestData(rootHash, false)
	}()

	w.WriteHeader(http.StatusOK)
}

func (s *Service) requestData(rootHash model.RootHash, isRetry bool) {
	_, err := s.db.GetRootHash(rootHash.Key())
	if err == nil {
		fmt.Printf("received root hash with key: %v already exist \n", rootHash.Key())
		return
	}
	if err != errors.ErrCannotFindRootHash {
		fmt.Print(err.Error())
		return
	}

	if err := s.db.SaveRootHash(rootHash); err != nil {
		fmt.Printf("saving root hash with key: %v failed due to error: %v", rootHash.Key(), err)
		return
	}

	client := dmsghttp.DMSGClient(s.config.Discovery, s.config.PubKey, s.config.SecKey)

	if err := s.retrieveHeaders(client, rootHash, rootHash.ObjectHeaderHash); err != nil {
		fmt.Printf("retrieveing headers failed due to error: %v", err)
		return
	}

	parcel, isValid := s.checkSignature(rootHash)
	s.db.RemoveUnreferencedObjects(rootHash.Key(), isValid)

	if !isValid && !isRetry {
		s.requestData(rootHash, true)
		return
	}

	if !isValid {
		fmt.Printf("Signature is not valid. All data from feed: %s is removed...", rootHash.Publisher)
		return
	}
	s.notifyRegisteredApps(rootHash, parcel)
	fmt.Println("Retrieving new data finished successfully")
}

func (s *Service) notifyRegisteredApps(rootHash model.RootHash, parcel model.Parcel) {
	addresses, err := s.db.GetAllRegisteredApps()
	if err != nil {
		fmt.Println("Error while fetching registered apps: ", err)
		return
	}
	notifyRequest := model.NotifyAppRequest{
		RootHash: rootHash,
		Parcel:   parcel,
	}
	b := new(bytes.Buffer)
	json.NewEncoder(b).Encode(notifyRequest)

	client := http.DefaultClient
	for _, address := range addresses {
		_, err := client.Post(fmt.Sprint("http://", address), "application/json", b)
		if err != nil {
			fmt.Printf("Notify app with address: %v failed due to error %v: ", address, err)
			continue
		}
		fmt.Printf("App with address: %v notified succesfully.", address)
	}
}

func (s *Service) retrieveHeaders(client *http.Client, rootHash model.RootHash, headerHashes ...string) error {
	headers, err := s.fetchObjectHeaders(client, headerHashes...)
	if err != nil {
		return fmt.Errorf("fetching object headers with hashes: %v from service failed due to error: %v", headerHashes, err)
	}
	var missingHeaderHashes []string
	for i, header := range headers {
		for _, ref := range header.ExternalReferences {
			existingHeader, err := s.db.GetObjectHeader(ref)
			if err != nil {
				if err == errors.ErrCannotFindObjectHeader {
					missingHeaderHashes = append(missingHeaderHashes, ref)
					continue
				}
				return fmt.Errorf("fetching object header with hash: %v from db failed due to error: %v", ref, err)
			} else {
				// update existing object header to newest sequence
				if err := s.UpdateObjectHeaderRootHashKey(ref, rootHash.Key(), existingHeader); err != nil {
					return err
				}
			}
		}
		// save missing object header
		if err := s.db.SaveObjectHeader(headerHashes[i], rootHash, header); err != nil {
			return fmt.Errorf("saving object header with hash: %v failed due to error: %v", headerHashes[i], err)
		}

		if len(header.ObjectHash) > 0 {
			// fetch and save missing object
			if err := s.fetchAndSaveObject(header.ObjectHash, headerHashes[i], client); err != nil {
				return err
			}
		}
	}

	if len(missingHeaderHashes) > 0 {
		if err := s.retrieveHeaders(client, rootHash, missingHeaderHashes...); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) fetchAndSaveObject(hash, objectHeaderHash string, client *http.Client) error {
	object, err := s.fetchObject(client, hash)
	if err != nil {
		return fmt.Errorf("fetch object with hash: %v failed due to error: %v", hash, err)
	}

	if err := s.db.SaveObject(hash, objectHeaderHash, object); err != nil {
		return fmt.Errorf("saving object with hash: %v failed due to error: %v", hash, err)
	}

	return nil
}

// Update object header and all references to the newest sequence
func (s *Service) UpdateObjectHeaderRootHashKey(hash, rootHashKey string, header model.ObjectHeader) error {
	if err := s.db.UpdateObjectHeaderRootHashKey(hash, rootHashKey); err != nil {
		return fmt.Errorf("updating object header with hash: %v failed due to error: %v", hash, err)
	}

	for _, ref := range header.ExternalReferences {
		header, err := s.db.GetObjectHeader(ref)
		if err != nil {
			return fmt.Errorf("retrieveing object header with hash: %v failed due to error: %v", ref, err)
		}
		if err := s.UpdateObjectHeaderRootHashKey(ref, rootHashKey, header); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) fetchObjectHeaders(client *http.Client, objectHeaderHashes ...string) ([]model.ObjectHeader, error) {
	objectHeadersResp := model.GetObjectHeadersResponse{}

	baseUrl := fmt.Sprint(s.config.TrackerAddress, "/data/object/header?hash=", objectHeaderHashes[0])
	additionalParams := ""
	for _, hash := range objectHeaderHashes[1:] {
		additionalParams = fmt.Sprint(additionalParams, "&hash=", hash)
	}

	url := fmt.Sprint(baseUrl, additionalParams)
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

func (s *Service) checkSignature(rootHash model.RootHash) (model.Parcel, bool) {
	parcel := model.Parcel{}
	s.recreateParcel(&parcel, rootHash.ObjectHeaderHash)

	parcelBytes, err := json.Marshal(parcel)
	if err != nil {
		panic(err)
	}

	sig := dmsgcipher.Sig{}
	_ = sig.UnmarshalText([]byte(rootHash.Signature))

	pubKey := dmsgcipher.PubKey{}
	_ = pubKey.UnmarshalText([]byte(rootHash.Publisher))

	if err = dmsgcipher.VerifyPubKeySignedPayload(pubKey, sig, parcelBytes); err != nil {
		fmt.Printf("parcel signature verification failed due to error: %v \n", err)
		return parcel, false
	}
	return parcel, true
}

func (s *Service) recreateParcel(parcel *model.Parcel, hash string) {
	header, err := s.db.GetObjectHeader(hash)
	if err != nil {
		panic(err)
	}
	parcel.ObjectHeaders = append(parcel.ObjectHeaders, header)

	if len(header.ObjectHash) == 0 {
		for _, ref := range header.ExternalReferences {
			s.recreateParcel(parcel, ref)
		}

	} else {
		object, err := s.db.GetObject(header.ObjectHash)
		if err != nil {
			panic(err)
		}

		parcel.Objects = append(parcel.Objects, object)
	}
}
