package node

import (
	"fmt"

	"github.com/SkycoinPro/cxo-2-node/src/node/database/boltdb"
	bolt "go.etcd.io/bbolt"
)

type data interface {
	saveObject(hash, path string) error
	isSaved(hash string) bool
}

type store struct {
	db *bolt.DB
}

func defaultData() data {
	return store{
		db: boltdb.DB,
	}
}

func (s store) saveObject(hash, path string) error {
	err := s.db.Update(func(tx *bolt.Tx) error {
		err := tx.Bucket([]byte(boltdb.DataObjectBucket)).Put([]byte(hash), []byte(path))
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
		b := tx.Bucket([]byte(boltdb.DataObjectBucket))
		pathBytes = b.Get([]byte(hash))
		return nil
	})

	return nil != pathBytes
}
