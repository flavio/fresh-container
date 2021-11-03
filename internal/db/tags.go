package db

import (
	"encoding/json"
	"time"

	"github.com/flavio/fresh-container/pkg/fresh_container"

	badger "github.com/dgraph-io/badger/v2"
	log "github.com/sirupsen/logrus"
)

func (d *DB) GetImageTags(image fresh_container.Image) ([]string, error) {
	var tags []string

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(image.FullNameWithoutTag() + image.TagPrefix))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return item.Value(func(val []byte) error {
			return json.Unmarshal(val, &tags)
		})
	})

	if err != nil {
		log.WithFields(log.Fields{
			"image": image.FullNameWithoutTag(),
			"error": err,
		}).Error("db.GetImageTags")
		return []string{}, err
	}

	log.WithFields(log.Fields{
		"image": image.FullNameWithoutTag(),
		"tags":  tags,
	}).Debug("db.GetImageTags")
	return tags, nil
}

func (d *DB) SetImageTags(image fresh_container.Image, tags []string) error {
	marshalledTags, err := json.Marshal(tags)
	if err != nil {
		log.WithFields(log.Fields{
			"image": image.FullNameWithoutTag(),
			"tags":  tags,
			"error": err,
		}).Error("db.SetImageTags")

		return err
	}

	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(image.FullNameWithoutTag()+image.TagPrefix), marshalledTags).
			WithTTL(time.Duration(d.config.CacheTTLHours) * time.Hour)
		return txn.SetEntry(entry)
	})
}
