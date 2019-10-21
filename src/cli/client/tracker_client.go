package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
)

type TrackerClient struct {
	client           *http.Client
	trackerAddress   string
	subscribeAddress string
}

func NewTrackerClient(cfg config.Config) *TrackerClient {
	return &TrackerClient{
		client:           dmsghttp.DMSGClient(cfg.Discovery, cfg.PubKey, cfg.SecKey),
		trackerAddress:   cfg.TrackerAddress,
		subscribeAddress: fmt.Sprintf("%v:%v/notify", cfg.PubKey, cfg.Port), //FIXME - read route from node service
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
	req.Header.Set("Address", t.subscribeAddress)

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("subscribe request failed due to error: %v", err)
	}

	fmt.Println("Subscribe response status: ", resp.Status)
	return nil
}

func (t *TrackerClient) SaveData(filePath string) error {
	bs, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("reading file: %v failed with error: %v", filePath, err)
	}
	r := bytes.NewReader(bs)
	url := fmt.Sprint(t.trackerAddress, saveDataRoute)
	req, err := http.NewRequest("POST", url, r)
	if err != nil {
		return fmt.Errorf("error creating new save data request for file: %v", filePath)
	}

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("save data request failed due to error: %v", err)
	}

	fmt.Println("Save data request response: ", resp.Status)
	return nil
}
