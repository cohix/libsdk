package driversqlite

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/cohix/libsdk/pkg/store"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"

	// using the "standard" driver that requires CGO, as the pure Go impl
	// available (github.com/glebarez/go-sqlite) had an inpenetrable error
	// "bad parameter or other API misuse: not an error (21)"
	// but will revisit when there's more time to debug and we have
	// to contend with the fun world of cross-compiling GCO.
	_ "github.com/mattn/go-sqlite3"
)

var _ store.Driver = &Sqlite{}

// Sqlite is a SQLite driver for libsdk store
type Sqlite struct {
	db  *sqlx.DB
	log slog.Logger
}

type Tx struct {
	tx       *sqlx.Tx
	didWrite bool
}

// ReadTx is a read-only transaction
type ReadTx struct {
	tx *sqlx.Tx
}

// ReadWriteTx is a read-write transaction
type ReadWriteTx struct {
	ReadTx
}

// New creates a new SQlite database on disk and a driver instance wrapping it.
func New(serviceName string) (store.Driver, error) {
	filepath, err := dbPath(serviceName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to dbPath")
	}

	db, err := sqlx.Connect("sqlite3", fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)", filepath))
	if err != nil {
		return nil, errors.Wrapf(err, "failed to sqlx.Connect for path %s", filepath)
	}

	s := &Sqlite{
		db:  db,
		log: *slog.With("lib", "libsdk", "pkg", "driversqlite"),
	}

	s.log.Info("database created", "file", filepath)

	return s, nil
}

// Exec executes a replayed transaction and returns its results
func (s *Sqlite) Exec(rec store.TxRecord, handler store.TxHandler) (store.Tx, any, error) {
	s.log.Debug(fmt.Sprintf("exec name:%s uuid:%s", rec.Name, rec.UUID))

	tx := s.tx()

	result, err := handler(tx, rec.Args...)
	if err != nil {
		if rbErr := tx.tx.Rollback(); rbErr != nil {
			return nil, nil, errors.Wrapf(rbErr, "failed to tx.Rollback after handler err %s", err.Error())
		} else {
			return nil, nil, errors.Wrap(err, "rolled back after error from handler")
		}
	}

	if err := tx.tx.Commit(); err != nil {
		return nil, nil, errors.Wrap(err, "failed to tx.Commit")
	}

	return tx, result, err
}

// Migrate runs migration statements that must all succeed or an error is returned
func (s *Sqlite) Migrate(statements []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return errors.Wrap(err, "failed to db.Begin")
	}

	for i, stmt := range statements {
		s.log.Info("running migration", "num", i, "of", len(statements))

		if _, err := tx.Exec(stmt); err != nil {
			return errors.Wrap(err, "failed to tx.Exec")
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to tx.Commit")
	}

	return nil
}

// tx returns a driver-compatible transaction
func (s *Sqlite) tx() *Tx {
	sqlxtx := s.db.MustBegin()

	t := &Tx{
		tx:       sqlxtx,
		didWrite: false,
	}

	return t
}

// Read returns a read-only transaction
func (t *Tx) Read() store.ReadTx {
	r := &ReadTx{
		tx: t.tx,
	}

	return r
}

// ReadWrite returns a read-write transaction
func (t *Tx) ReadWrite() store.ReadWriteTx {
	t.didWrite = true

	rw := &ReadWriteTx{
		ReadTx{
			tx: t.tx,
		},
	}

	return rw
}

// DidWrite returns true if the transaction was used to write.
func (t *Tx) DidWrite() bool {
	return t.didWrite
}

// Select runs a query to select one or more rows and read them into out.
func (r *ReadTx) Select(out any, query string, args ...any) error {
	if err := r.tx.Select(out, query, args...); err != nil {
		return errors.Wrap(err, "failed to tx.Select")
	}

	return nil
}

// Get runs a query to select a single row and read it into out.
func (r *ReadTx) Get(out any, query string, args ...any) error {
	if err := r.tx.Get(out, query, args...); err != nil {
		return errors.Wrap(err, "failed to tx.Get")
	}

	return nil
}

// Exec runs the provided query with the provided args and returns the insert ID, if any
// Exec should be used for any insert, update, or delete queries.
func (rw *ReadWriteTx) Exec(query string, args ...any) (int64, error) {
	result, err := rw.tx.Exec(query, args...)
	if err != nil {
		return 0, errors.Wrap(err, "failed to tx.Exec")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, "failed to LastInsertID")
	}

	return id, nil
}

func dbPath(serviceName string) (string, error) {
	config, err := os.UserCacheDir()
	if err != nil {
		return "", errors.Wrap(err, "failed to UserCacheDir")
	}

	folder := fmt.Sprintf("%s/libsdk/%s", config, serviceName)

	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		return "", errors.Wrap(err, "failed to MkdirAll")
	}

	dbUUID, err := uuid.NewV7()
	if err != nil {
		return "", errors.Wrap(err, "failed to uuid.NewV7")
	}

	// each time the service starts up, it's going to re-create the db from scratch
	// by replaying from the fabric, so each time we create a new db to ensure it's fresh
	dir := fmt.Sprintf("%s/%s-%s.sqlite", folder, serviceName, dbUUID.String())

	return dir, nil
}
