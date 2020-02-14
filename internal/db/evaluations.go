package db

import (
	"fmt"
	"time"

	badger "github.com/dgraph-io/badger/v2"
)

func (d *DB) SetEvaluation(id string, result []byte) error {
	key := fmt.Sprintf("evaluations/%s", id)
	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), result).
			WithTTL(time.Duration(d.config.CacheTTLHours) * time.Hour)
		return txn.SetEntry(entry)
	})
}

func (d *DB) GetEvaluation(id string) (string, error) {
	key := fmt.Sprintf("evaluations/%s", id)
	var result string

	err := d.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				return ErrorEvaluationNotFound
			}
			return err
		}

		return item.Value(func(val []byte) error {
			result = string(val)
			return nil
		})
	})

	if err != nil {
		return "", err
	}

	return result, nil
}
