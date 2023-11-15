package service

import (
	"net/http"
	"os"

	"github.com/cohix/libsdk/pkg/fabric"
	fabricnats "github.com/cohix/libsdk/pkg/fabric/fabric-nats"
	"github.com/cohix/libsdk/pkg/store"
	driversqlite "github.com/cohix/libsdk/pkg/store/driver-sqlite"
	"github.com/pkg/errors"
)

const publicAddrEnvKey = "LIBSDK_PUBLIC_ADDR"

// Service is a libsdk service which contains public and private servers,
// a fabric, and a store for simple service development.
type Service struct {
	name string

	fabric fabric.Fabric
	store  *store.Store
}

// New creates a Service with a SQLite store and NATS fabric.
func New(name string) (*Service, error) {
	f, err := fabricnats.New(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fabricnats.New")
	}

	r, err := f.Replayer("store", true)
	if err != nil {
		return nil, errors.Wrap(err, "failed to f.Replayer")
	}

	d, err := driversqlite.New(name)
	if err != nil {
		return nil, errors.Wrap(err, "failed to driversqlite.New")
	}

	s := store.New(d, r)

	return NewWithFabricStore(name, f, s)
}

// NewWithFabricStore creates a Service with the provided fabric and store.
func NewWithFabricStore(name string, fabric fabric.Fabric, store *store.Store) (*Service, error) {
	s := &Service{
		name:   name,
		fabric: fabric,
		store:  store,
	}

	return s, nil
}

// Serve takes in an App definition and begins serving the public and private handlers.
// - App's public handler is served on an HTTP port defined by LIBSDK_PUBLIC_PORT, defaulting to :8080
// - App's private handler is served using the configured fabric.
// - App's transaction handlers are registered for use by the store.
// - App's migrations are applied to the store before replaying transactions.
func (s *Service) Serve(app App) error {
	for name, handler := range app.Transactions() {
		if err := s.store.Register(name, handler); err != nil {
			return errors.Wrap(err, "failed to store.Register")
		}
	}

	if err := s.store.Start(app.Migrations()); err != nil {
		return errors.Wrap(err, "failed to store.Start")
	}

	server := &http.Server{
		Addr:    publicAddr(),
		Handler: app.Public(s.store),
	}

	app.Log().Info("public server starting", "addr", server.Addr)

	return server.ListenAndServe()
}

// Store returns the Service's store, which should be used by handlers
// to read and write from the replicated database via store.Exec.
func (s *Service) Store() *store.Store {
	return s.store
}

func publicAddr() string {
	addr := ":8080"

	envAddr, exists := os.LookupEnv(publicAddrEnvKey)
	if exists {
		addr = envAddr
	}

	return addr
}
