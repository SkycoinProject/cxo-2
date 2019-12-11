package data

import (
	"time"

	"github.com/SkycoinPro/cxo-2-node/pkg/errors"

	"github.com/SkycoinPro/cxo-2-node/pkg/model"
	storm "github.com/asdine/storm/v3"
	"github.com/asdine/storm/v3/q"
	log "github.com/sirupsen/logrus"
)

type Data interface {
	SaveRootHash(rootHash model.RootHash) error
	SaveObjectHeader(hash string, rootHash model.RootHash, objectHeader model.ObjectHeader) error
	SaveObjectInfo(hash, path string) error
	UpdateObjectHeaderRootHashKey(hash string, rootHashKey string) error
	GetRootHash(hash string) (model.RootHash, error)
	GetObjectHeader(hash string) (model.ObjectHeader, error)
	FindNewObjectHeaderHashes(rootHashKey string, timestamp time.Time) (map[string]struct{}, error)
}

type store struct {
	db *storm.DB
}

func DefaultData() Data {
	return store{
		db: DB,
	}
}

func (s store) SaveRootHash(rootHash model.RootHash) error {
	return s.db.Save(&rootHashDAO{
		ID:       rootHash.Key(),
		RootHash: rootHash,
	})
}

func (s store) SaveObjectHeader(hash string, rootHash model.RootHash, objectHeader model.ObjectHeader) error {
	return s.db.Save(&objectHeaderDAO{
		ID:           hash,
		RootHashKey:  rootHash.Key(),
		Timestamp:    rootHash.Timestamp,
		ObjectHeader: objectHeader,
	})
}

func (s store) SaveObjectInfo(hash, path string) error {
	return s.db.Save(&objectInfo{
		ID:   hash,
		Path: path,
	})
}

func (s store) UpdateObjectHeaderRootHashKey(hash string, rootHashKey string) error {
	return s.db.UpdateField(&objectHeaderDAO{ID: hash}, "RootHashKey", rootHashKey)
}

func (s store) GetRootHash(key string) (model.RootHash, error) {
	rootHashDAO := rootHashDAO{}
	var err error
	if dbError := s.db.One("ID", key, &rootHashDAO); dbError != nil {
		if dbError == storm.ErrNotFound {
			err = errors.ErrCannotFindRootHash
		} else {
			log.Errorf("could not retrieve root hash with key: %v due to error: %v", key, err)
			err = dbError
		}
	}

	return rootHashDAO.RootHash, err
}

func (s store) GetObjectHeader(hash string) (model.ObjectHeader, error) {
	objectHeaderDAO := objectHeaderDAO{}
	var err error
	if dbError := s.db.One("ID", hash, &objectHeaderDAO); dbError != nil {
		if dbError == storm.ErrNotFound {
			err = errors.ErrCannotFindObjectHeader
		} else {
			log.Errorf("could not retrieve object header with hash: %v due to error: %v", hash, err)
			err = dbError
		}
	}

	return objectHeaderDAO.ObjectHeader, err
}

func (s store) FindNewObjectHeaderHashes(rootHashKey string, timestamp time.Time) (map[string]struct{}, error) {
	var objectHeaderDAOs []objectHeaderDAO
	if err := s.db.Select(q.Eq("RootHashKey", rootHashKey), q.Eq("Timestamp", timestamp)).Find(&objectHeaderDAOs); err != nil {
		if err == storm.ErrNotFound {
			return make(map[string]struct{}, 0), nil
		}
		log.Errorf("could not retrieve object headers with root hash key: %v and timestamp %v due to error: %v", rootHashKey, timestamp, err)
		return make(map[string]struct{}, 0), err
	}

	headerHashes := make(map[string]struct{}, len(objectHeaderDAOs))
	for _, dao := range objectHeaderDAOs {
		headerHashes[dao.ID] = struct{}{}
	}

	return headerHashes, nil
}
