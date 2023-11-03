package main

import (
	"log/slog"
	"net/http"

	"github.com/cohix/libsdk/pkg/service"
	"github.com/cohix/libsdk/pkg/store"
)

type RolesApp struct {
	log *slog.Logger
}

var _ service.App = &RolesApp{}

// Migrations returns the app's DB migrations.
func (p *RolesApp) Migrations() []string {
	return []string{}
}

// Transactions returns the transactions available to the app.
func (p *RolesApp) Transactions() map[string]store.TxHandler {
	txs := map[string]store.TxHandler{}

	return txs
}

// Public returns the HTTP router for public-facing requests.
func (p *RolesApp) Public(store *store.Store) http.Handler {
	mux := http.NewServeMux()

	return mux
}

// Private returns the HTTP router for private inter-service requests.
func (p *RolesApp) Private(store *store.Store) http.Handler {
	return http.NewServeMux()
}

// Log returns a logger configured to the preferences of the app.
func (p *RolesApp) Log() *slog.Logger {
	return p.log
}
