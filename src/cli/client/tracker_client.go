package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	"github.com/SkycoinPro/cxo-2-node/src/model"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
	"github.com/skycoin/dmsg/cipher"
)

type TrackerClient struct {
	client           *http.Client
	trackerAddress   string
	subscribeAddress string
}

func NewTrackerClient(cfg config.Config) *TrackerClient {
	sPK, sSK := cipher.GenerateKeyPair()
	return &TrackerClient{
		client:           dmsghttp.DMSGClient(cfg.Discovery, sPK, sSK),
		trackerAddress:   cfg.TrackerAddress,
		subscribeAddress: fmt.Sprintf("%v:%v/notify?hash=", cfg.PubKey.Hex(), cfg.Port), //FIXME - read route from node service
	}
}

const (
	subscribeRoute = "/subscribe?pubKey="
	saveDataRoute  = "/data"
)

func (t *TrackerClient) Subscribe(publicKey string) error {
	url := fmt.Sprint(t.trackerAddress, subscribeRoute, publicKey)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("error creating subscribe request to public key: %v ", publicKey)
	}

	fmt.Println("subscribe address: ", t.subscribeAddress)
	req.Header.Set("Address", t.subscribeAddress)

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe request failed due to error: %v", err)
	}

	fmt.Println("Subscribe response status: ", resp.Status)
	return nil
}

func (t *TrackerClient) SaveData(request model.PublishDataRequest) error {
	bs, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal request failed due to error: %v", err)
	}
	r := bytes.NewReader(bs)
	url := fmt.Sprint(t.trackerAddress, saveDataRoute)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return fmt.Errorf("creating publish new data request failed due to error:%v", err)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("publish data request failed due to error: %v", err)
	}

	cipher.GenerateKeyPair()

	fmt.Println("Publish data request response: ", resp.Status)
	return nil
}
