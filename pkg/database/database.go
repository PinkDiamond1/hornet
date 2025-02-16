package database

import (
	"encoding/json"
	"time"

	"github.com/pkg/errors"

	"github.com/iotaledger/hive.go/core/events"
	"github.com/iotaledger/hive.go/core/ioutils"
	"github.com/iotaledger/hive.go/core/kvstore"
	"github.com/iotaledger/hornet/v2/pkg/metrics"
)

type Engine string

const (
	EngineUnknown Engine = "unknown"
	EngineAuto    Engine = "auto"
	EngineRocksDB Engine = "rocksdb"
	EnginePebble  Engine = "pebble"
	EngineMapDB   Engine = "mapdb"
)

var (
	// ErrNothingToCleanUp is returned when nothing is there to clean up in the database.
	ErrNothingToCleanUp = errors.New("Nothing to clean up in the databases")
)

type Cleanup struct {
	Start time.Time
	End   time.Time
}

func (c *Cleanup) MarshalJSON() ([]byte, error) {

	cleanup := struct {
		Start int64 `json:"start"`
		End   int64 `json:"end"`
	}{
		Start: 0,
		End:   0,
	}

	if !c.Start.IsZero() {
		cleanup.Start = c.Start.Unix()
	}

	if !c.End.IsZero() {
		cleanup.End = c.End.Unix()
	}

	return json.Marshal(cleanup)
}

func CleanupCaller(handler interface{}, params ...interface{}) {
	//nolint:forcetypeassert // we will replace that with generic events anyway
	handler.(func(*Cleanup))(params[0].(*Cleanup))
}

type Events struct {
	Cleanup    *events.Event
	Compaction *events.Event
}

// Database holds the underlying KVStore and database specific functions.
type Database struct {
	databaseDir           string
	store                 kvstore.KVStore
	engine                Engine
	metrics               *metrics.DatabaseMetrics
	events                *Events
	compactionSupported   bool
	compactionRunningFunc func() bool
}

// New creates a new Database instance.
func New(databaseDirectory string, kvStore kvstore.KVStore, engine Engine, metrics *metrics.DatabaseMetrics, events *Events, compactionSupported bool, compactionRunningFunc func() bool) *Database {
	return &Database{
		databaseDir:           databaseDirectory,
		store:                 kvStore,
		engine:                engine,
		metrics:               metrics,
		events:                events,
		compactionSupported:   compactionSupported,
		compactionRunningFunc: compactionRunningFunc,
	}
}

// KVStore returns the underlying KVStore.
func (db *Database) KVStore() kvstore.KVStore {
	return db.store
}

// Engine returns the database engine.
func (db *Database) Engine() Engine {
	return db.engine
}

// Metrics returns the database metrics.
func (db *Database) Metrics() *metrics.DatabaseMetrics {
	return db.metrics
}

// Events returns the events of the database.
func (db *Database) Events() *Events {
	return db.events
}

// CompactionSupported returns whether the database engine supports compaction.
func (db *Database) CompactionSupported() bool {
	return db.compactionSupported
}

// CompactionRunning returns whether a compaction is running.
func (db *Database) CompactionRunning() bool {
	if db.compactionRunningFunc == nil {
		return false
	}

	return db.compactionRunningFunc()
}

// Size returns the size of the database.
func (db *Database) Size() (int64, error) {
	if db.engine == EngineMapDB {
		// in-memory database does not support this method.
		return 0, nil
	}

	return ioutils.FolderSize(db.databaseDir)
}
