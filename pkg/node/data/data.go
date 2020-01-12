package data

import (
	"strings"
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
	GetObjectPath(hash string) (string, error)
	FindNewObjectHeaderHashes(rootHashKey string, timestamp time.Time) (map[string]struct{}, error)
	RemoveUnreferencedObjects(rootHashKey string, isValidSignature bool) []string
	RegisterApp(address, name string) error
	GetAllRegisteredApps() ([]string, error)
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
	var info objectInfo
	if err := s.db.One("Path", path, &info); err != nil {
		if err != storm.ErrNotFound {
			log.Errorf("could not find previous object info with path: %v due to error: %v", path, err)
			return err
		}
	}

	if info.ID != "" {
		// delete previous record for same path in order not to delete path on clean up later on
		if err := s.db.DeleteStruct(&info); err != nil {
			log.Errorf("could not delete previous object info with path: %v due to error: %v", path, err)
			return err
		}
	}

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

func (s store) GetObjectPath(hash string) (string, error) {
	objectInfo := objectInfo{}
	var err error
	if dbError := s.db.One("ID", hash, &objectInfo); dbError != nil {
		if dbError == storm.ErrNotFound {
			err = errors.ErrCannotFindObjectPath
		} else {
			log.Errorf("could not retrieve object path with hash: %v due to error: %v", hash, err)
			err = dbError
		}
	}

	return objectInfo.Path, err
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

// Remove headers and object infos and return slice of paths for dirs and files that should be removed
func (s store) RemoveUnreferencedObjects(latestRootHashKey string, isValidSignature bool) []string {
	var objectHeaderDAOs []objectHeaderDAO
	pubKey := strings.Split(latestRootHashKey, "_")[0]

	//selecting all headers for specific pub key
	if err := s.db.Prefix("RootHashKey", pubKey, &objectHeaderDAOs); err != nil {
		if err != storm.ErrNotFound {
			log.Errorf("could not retrieve unreferenced object headers due to error: %v", err)
			return []string{}
		}
	}

	var unreferencedObjectHeaderDAOs []objectHeaderDAO
	// if signature is valid we are removing every object that is not on latest sequence
	if isValidSignature {
		for _, h := range objectHeaderDAOs {
			if latestRootHashKey != h.RootHashKey {
				unreferencedObjectHeaderDAOs = append(unreferencedObjectHeaderDAOs, h)
			}
		}
	} else {
		// if signature is not valid we are removing all objects for specific pub key
		unreferencedObjectHeaderDAOs = objectHeaderDAOs
	}

	var paths []string
	for _, headerDAO := range unreferencedObjectHeaderDAOs {
		var objectInfoID string
		if isDirectory(headerDAO.ObjectHeader) {
			objectInfoID = headerDAO.ID
		} else {
			objectInfoID = headerDAO.ObjectHeader.ObjectHash
		}

		var objectInfo objectInfo
		if err := s.db.One("ID", objectInfoID, &objectInfo); err != nil {
			if err != storm.ErrNotFound {
				log.Errorf("Fetching object info with hash: %v failed with error: %v", headerDAO.ID, err)
			}
		} else {
			paths = append(paths, objectInfo.Path)
			if err := s.db.DeleteStruct(&objectInfo); err != nil {
				log.Errorf("Deleting object info with hash: %v failed with error: %v", objectInfo.ID, err)
			}

		}
		if err := s.db.DeleteStruct(&headerDAO); err != nil {
			log.Errorf("Deleting object header with hash: %v failed with error: %v", headerDAO.ID, err)
		}
	}

	if !isValidSignature {
		rootHashDAO := rootHashDAO{}
		if err := s.db.One("ID", latestRootHashKey, &rootHashDAO); err != nil {
			log.Errorf("could not retrieve root hash with key: %v due to error: %v", latestRootHashKey, err)
		} else {
			_ = s.db.DeleteStruct(&rootHashDAO)
		}
	}

	return paths
}

func isDirectory(oh model.ObjectHeader) bool {
	for _, meta := range oh.Meta {
		if meta.Key == "type" && meta.Value == "directory" {
			return true
		}
	}
	return false
}

func (s store) RegisterApp(address, name string) error {
	var existingApp app
	if err := s.db.One("Address", address, &existingApp); err != nil {
		if err != storm.ErrNotFound {
			log.Errorf("fetching app with address: %v failed due to error: %v", address, err)
			return err
		}
	} else {
		log.Infof("app with address: %v already registered...", address)
		return nil
	}
	return s.db.Save(&app{
		Address: address,
		Name:    name,
	})
}

func (s store) GetAllRegisteredApps() ([]string, error) {
	var apps []app
	var addresses []string
	var err error
	if err = s.db.All(&apps); err != nil {
		log.Error("could not retrieve registered apps due to error: ", err)
		return addresses, err
	}

	for _, app := range apps {
		addresses = append(addresses, app.Address)
	}

	return addresses, nil
}
