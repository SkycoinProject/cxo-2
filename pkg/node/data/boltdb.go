package data

import (
	"fmt"

	storm "github.com/asdine/storm/v3"

	log "github.com/sirupsen/logrus"
)

// DB global store variable
var DB *storm.DB

const (
	databaseName = "cxo-node.db"
)

// Init - initialization of bolt db
func Init() func() {
	var err error
	DB, err = storm.Open(databaseName)
	if err != nil {
		log.Fatal("Failed to connect to database", err)
	}

	log.Info("Database connected. Checking buckets...")
	if err = initBuckets(); err != nil {
		log.Fatal("buckets initialization failed due to error: ", err)
	}

	return func() {
		log.Info("Disconnecting database")
		if err := DB.Close(); err != nil {
			log.Fatal("closing db failed due to error: ", err)
		}
		log.Debug("Database disconnected")
	}
}

func initBuckets() error {
	err := DB.Init(&rootHashDAO{})
	if err != nil {
		return fmt.Errorf("could not create root hash bucket: %v", err)
	}

	err = DB.Init(&objectHeaderDAO{})
	if err != nil {
		return fmt.Errorf("could not create object header bucket: %v", err)
	}

	err = DB.Init(&objectInfo{})
	if err != nil {
		return fmt.Errorf("could not create object bucket: %v", err)
	}

	err = DB.Init(&app{})
	if err != nil {
		return fmt.Errorf("could not create app bucket: %v", err)
	}

	return nil
}
