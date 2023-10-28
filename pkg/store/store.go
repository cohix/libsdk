package store

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/cohix/libsdk/pkg/fabric"
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
)

// Store is a distributed, replicated datastore for libsdk applications
type Store struct {
	driver       Driver
	replayer     fabric.ReplayConnection
	log          *slog.Logger
	transactions map[string]TxHandler
	inflight     sync.Map
}

// Driver represents an underlying storage driver
type Driver interface {
	Exec(record TxRecord, handler TxHandler) (tx Tx, result any, err error)
	Migrate(statements []string) error
}

// Tx is an object that can itself kick off a read-only transaction
// or a read-write transaction which is managed by the underlying driver
type Tx interface {
	Read() ReadTx
	ReadWrite() ReadWriteTx
	DidWrite() bool
}

// ReadTx is a read-only transaction
type ReadTx interface {
	Select(out []any, query string, args ...any) error
	Get(out any, query string, args ...any) error
}

// ReadWriteTx is a transaction for read or write transactions
type ReadWriteTx interface {
	ReadTx
	Exec(query string, args ...any) (int64, error)
	Delete(query string, args ...any) error
}

// TxHandler is a function that executes a named transaction
type TxHandler func(tx Tx, args ...any) (result any, err error)

// TxRecord is a serializable transaction for replication purposes
type TxRecord struct {
	UUID string `json:"uuid"`
	Name string `json:"name"`
	Args []any  `json:"args"`
}

// New creates a new Store with the given driver
func New(driver Driver, replayer fabric.ReplayConnection) *Store {
	s := &Store{
		driver:       driver,
		replayer:     replayer,
		log:          slog.With("lib", "libsdk", "module", "store"),
		transactions: map[string]TxHandler{},
		inflight:     sync.Map{},
	}

	return s
}

// Start starts the store replay loop
func (s *Store) Start(migrations []string) error {
	if err := s.driver.Migrate(migrations); err != nil {
		return errors.Wrap(err, "failed to driver.Migrate")
	}

	msgGenerator := func() any {
		return &TxRecord{}
	}

	msgHandler := func(msg any) {
		txRec := msg.(*TxRecord)

		s.log.Debug("replaying transaction", "uuid", txRec.UUID)

		handler, exists := s.transactions[txRec.Name]
		if !exists {
			s.log.Error(fmt.Sprintf("named transaction %s is not registered", txRec.Name))
			return
		}

		completion, exists := s.inflight.LoadAndDelete(txRec.UUID)
		// if this is a new, in-flight transaction, it's already been executed,
		// so we call its completion func to let the caller know it's done and exit
		if exists {
			cmplFunc := completion.(context.CancelFunc)
			cmplFunc()
			return
		}

		_, _, err := s.driver.Exec(*txRec, handler)
		if err != nil {
			s.log.Error(errors.Wrapf(err, "failed to Exec replayed transaction %s with name %s", txRec.UUID, txRec.Name).Error())
		}
	}

	// Replay will continue async even after the upToDate channel
	// fires, but once it does, it is safe to continue as the db is
	// up to date and ready for new queries etc.
	upToDate, err := s.replayer.Replay(msgGenerator, msgHandler)
	if err != nil {
		s.log.Error(errors.Wrap(err, "failed to replayer.Replay").Error())
	}

	<-upToDate

	return nil
}

// Register registers the given transaction under the given name.
// name must be unique, attempt to re-register with same name results in an error.
func (s *Store) Register(name string, handler TxHandler) error {
	_, exists := s.transactions[name]
	if exists {
		return fmt.Errorf("transaction registered with name %s already exists", name)
	}

	s.transactions[name] = handler

	return nil
}

// Exec performs a two-stage distributed transaction based on a
// registered named TxHandler. The Tx is distributed using the fabric
// and, upon confirmation of successful distribution, applied to the
// local store replica. The result or error of the TxHandler is returned.
// Non-errored call to Exec guarantees that replication succeeded.
func (s *Store) Exec(name string, args ...any) (any, error) {
	handler, exists := s.transactions[name]
	if !exists {
		return nil, fmt.Errorf("transaction with name %s is not registered", name)
	}

	txUUID, err := uuid.NewV7()
	if err != nil {
		return nil, errors.Wrap(err, "failed to uuid.NewV7")
	}

	txRec := TxRecord{
		UUID: txUUID.String(),
		Name: name,
		Args: args,
	}

	// by this point, the driver has already either committed
	// or rolled back the transaction internally, but it's
	// returned so that we can determine if it should be distributed
	tx, result, err := s.driver.Exec(txRec, handler)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to Exec transaction %s with name %s", txRec.UUID, txRec.Name)
	}

	// if the transaction did not write, there is no
	// reason to distribute it, so return its result early
	if tx != nil && !tx.DidWrite() {
		return result, err
	}

	pubCtx, cancel := context.WithTimeout(context.Background(), time.Duration(time.Second*30))

	s.inflight.Store(txRec.UUID, cancel)

	if err := s.replayer.Publish(txRec); err != nil {
		return nil, errors.Wrapf(err, "failed to replayer.Publish for tx %s with name %s", txRec.UUID, txRec.Name)
	}

	<-pubCtx.Done()

	if err = pubCtx.Err(); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, errors.Wrap(err, "publish context timed out")
		}
	}

	// the completionFunc above sets these when the two-stage transaction returns
	return result, nil
}
