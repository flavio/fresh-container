package db

import (
	"fmt"
	"strings"
	"time"

	badger "github.com/dgraph-io/badger/v2"
)

func (d *DB) SetEvaluation(id string, result []byte, status string) error {
	key := fmt.Sprintf("evaluations/%s", id)
	return d.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), result).
			WithTTL(time.Duration(d.config.CacheTTLHours) * time.Hour)
		err := txn.SetEntry(entry)
		if err == nil && status != "" {
			key := fmt.Sprintf("evaluations-status/%s", id)
			entry := badger.NewEntry([]byte(key), []byte(status)).
				WithTTL(time.Duration(d.config.CacheTTLHours) * time.Hour)
			err = txn.SetEntry(entry)
		}
		return err
	})
}

func (d *DB) GetEvaluation(id string) (string, error) {
	key := fmt.Sprintf("evaluations/%s", id)
	statusKey := fmt.Sprintf("evaluations-status/%s", id)
	var result string
	var resultStatus string
	var errStatus error

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

	if result != "" {
		errStatus = d.db.View(func(txn *badger.Txn) error {
			item, err := txn.Get([]byte(statusKey))
			if err != nil {
				if err == badger.ErrKeyNotFound {
					return nil
				} else {
					return err
				}
			}
			return item.Value(func(val []byte) error {
				resultStatus = string(val)
				return nil
			})
		})
	}

	if err != nil {
		return "", err
	}
	if errStatus != nil {
		return "", errStatus
	}
	if resultStatus != "" {
		return "", fmt.Errorf("%s", strings.ReplaceAll(result, "\\", ""))
	}

	return result, nil
}
