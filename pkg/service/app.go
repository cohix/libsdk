package service

import (
	"log/slog"
	"net/http"

	"github.com/cohix/libsdk/pkg/store"
)

type AppHandlerFunc func(store *store.Store) http.Handler

// App provides an application's logic to a service.
type App interface {
	Migrations() []string
	Transactions() map[store.TxName]store.TxHandler
	Public(store *store.Store) http.Handler
	Private(store *store.Store) http.Handler
	Log() *slog.Logger
}

// simpleApp is the minimum required
type simpleApp struct {
	migrations    []string
	transactions  map[string]store.TxHandler
	publicHandler AppHandlerFunc
}

// SimpleApp returns a minimum viable App for use with Serve()
func SimpleApp(migrations []string, transactions map[string]store.TxHandler, handler AppHandlerFunc) *simpleApp {
	s := &simpleApp{
		migrations:    migrations,
		transactions:  transactions,
		publicHandler: handler,
	}

	return s
}

// Migrations returns the app's DB migrations.
func (s *simpleApp) Migrations() []string {
	return s.migrations
}

// Transactions returns the transactions available to the app.
func (s *simpleApp) Transactions() map[string]store.TxHandler {
	return s.transactions
}

// Public returns the HTTP router for public-facing requests.
func (s *simpleApp) Public(store *store.Store) http.Handler {
	return s.publicHandler(store)
}

// Private returns the HTTP router for private inter-service requests.
func (s *simpleApp) Private(store *store.Store) http.Handler {
	return http.NewServeMux()
}

// Log returns a logger configured to the preferences of the app.
func (s *simpleApp) Log() *slog.Logger {
	return slog.Default()
}
