package node

import (
	"encoding/json"
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/model"
	log "github.com/sirupsen/logrus"

	"github.com/SkycoinPro/cxo-2-node/src/node/database/boltdb"
	bolt "go.etcd.io/bbolt"
)

type data interface {
	saveObjectHeader(hash string, objectHeader model.ObjectHeader) error
	saveObject(hash, path string) error
	isSaved(hash string) bool
	getObjectHeader(hash string) (model.ObjectHeader, error)
}

type store struct {
	db *bolt.DB
}

func defaultData() data {
	return store{
		db: boltdb.DB,
	}
}

func (s store) saveObjectHeader(hash string, objectHeader model.ObjectHeader) error {
	objectHeaderBytes, err := json.Marshal(objectHeader)
	if err != nil {
		return fmt.Errorf("could not marshal object header: %v", err)
	}

	return s.db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte(boltdb.ObjectHeaderBucket)).Put([]byte(hash), objectHeaderBytes)
		if err != nil {
			return fmt.Errorf("saving object header with hash: %v failed due to error: %v", hash, err)
		}
		log.Info("Saved object header with hash: ", hash)
		return nil
	})
}

func (s store) saveObject(hash, path string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(boltdb.ObjectBucket)).Put([]byte(hash), []byte(path))
		if err != nil {
			return fmt.Errorf("saving object hash: %v failed due to error: %v", hash, err)
		}
		return nil
	})

	return err
}

func (s store) isSaved(hash string) bool {
	var pathBytes []byte
	_ = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltdb.ObjectHeaderBucket))
		pathBytes = b.Get([]byte(hash))
		return nil
	})

	return nil != pathBytes
}

func (s store) getObjectHeader(hash string) (model.ObjectHeader, error) {
	objectHeader := model.ObjectHeader{}
	var objectHeaderBytes []byte

	_ = s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(boltdb.ObjectHeaderBucket))
		objectHeaderBytes = b.Get([]byte(hash))
		return nil
	})

	if len(objectHeaderBytes) > 0 {
		if err := json.Unmarshal(objectHeaderBytes, &objectHeader); err != nil {
			return model.ObjectHeader{}, fmt.Errorf("could not unmarshal object header due to error: %v", err)
		}

		return objectHeader, nil
	}

	return objectHeader, errCannotFindObjectHeader
}
