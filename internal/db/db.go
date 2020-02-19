package db

import (
	"github.com/flavio/fresh-container/internal/config"

	badger "github.com/dgraph-io/badger/v2"
	"github.com/pkg/errors"
)

type DB struct {
	db     *badger.DB
	config *config.Config
}

var (
	ErrorEvaluationNotFound = errors.New("Evaluation not found")
)

func NewDB(config *config.Config) (*DB, error) {
	opt := badger.DefaultOptions("").WithInMemory(true)
	bdb, err := badger.Open(opt)

	if err != nil {
		return nil, err
	}

	return &DB{
		db:     bdb,
		config: config,
	}, nil
}

func (d *DB) GC() error {
	return d.db.RunValueLogGC(0.7)
}
