package boltdb

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	bolt "go.etcd.io/bbolt"
)

// DB global store variable
var DB *bolt.DB

const (
	databaseName = "objects.db"
	// ObjectHeaderBucket - object header bucket name
	ObjectHeaderBucket = "OBJECT_HEADERS"
	// ObjectBucket - object bucket name
	ObjectBucket = "OBJECTS"
)

// Init - initialization of bolt db
func Init() func() {
	var err error
	DB, err = bolt.Open(databaseName, 0600, nil)
	if err != nil {
		log.Fatal("Failed to connect to bolt database", err)
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
	return DB.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(ObjectBucket))
		if err != nil {
			return fmt.Errorf("could not create object bucket: %v", err)
		}
		_, err = tx.CreateBucketIfNotExists([]byte(ObjectHeaderBucket))
		if err != nil {
			return fmt.Errorf("could not create object header bucket: %v", err)
		}
		return nil
	})
}
