package node

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/SkycoinPro/cxo-2-node/src/config"
	dmsghttp "github.com/SkycoinProject/dmsg-http"
)

// Service - node service model
type Service struct {
	config config.Config
}

// NewService - initialize node service
func NewService(cfg config.Config) *Service {
	return &Service{config: cfg}
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

	// prepare server route handling
	mux := http.NewServeMux()
	mux.HandleFunc(notifyRoute, notifyHandler)

	// run the server
	sErr := make(chan error, 1)
	sErr <- httpS.Serve(mux)
	close(sErr)
}

//TODO - Add check for existing file and saving new file
func notifyHandler(_ http.ResponseWriter, r *http.Request) {
	var hash string
	if err := json.NewDecoder(r.Body).Decode(&hash); err != nil {
		fmt.Printf("Error while receiving new hash: %v from cxo tracker service", hash)
	}

	fmt.Printf("New root hash: %v from cxo tracker received", hash)
}
