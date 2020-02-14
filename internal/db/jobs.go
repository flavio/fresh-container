package db

import (
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v2"
)

func (d *DB) SetJobQueued(id string) error {
	key := fmt.Sprintf("jobs/%s", id)
	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), []byte("queued")).
			WithTTL(time.Duration(d.config.CacheTTLHours) * time.Hour)
		return txn.SetEntry(entry)
	})
}

func (d *DB) IsJobQueued(id string) (bool, error) {
	key := fmt.Sprintf("jobs/%s", id)
	queued := false

	err := d.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		queued = true
		return nil
	})

	if err != nil {
		return false, err
	}

	return queued, nil
}

func (d *DB) RemoveJobFromQueue(id string) error {
	key := fmt.Sprintf("jobs/%s", id)

	err := d.db.Update(func(txn *badger.Txn) error {
		err := txn.Delete([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			}
			return err
		}

		return nil
	})

	return err
}
