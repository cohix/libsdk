package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/cohix/libsdk/pkg/service"
	"github.com/cohix/libsdk/pkg/store"
	"github.com/pkg/errors"
)

type PersonApp struct {
	log *slog.Logger
}

var _ service.App = &PersonApp{}

// Migrations returns the app's DB migrations.
func (p *PersonApp) Migrations() []string {
	return personSvcMigrations()
}

// Transactions returns the transactions available to the app.
func (p *PersonApp) Transactions() map[string]store.TxHandler {
	txs := map[string]store.TxHandler{
		"InsertPerson": InsertPersonHandler,
		"SelectPeople": SelectPeopleHandler,
		"GetPerson":    GetPersonHandler,
	}

	return txs
}

// Public returns the HTTP router for public-facing requests.
func (p *PersonApp) Public(store *store.Store) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/insert", p.insertHandler(store))
	mux.HandleFunc("/select", p.selectHandler(store))
	mux.HandleFunc("/get", p.getHandler(store))

	return mux
}

// Private returns the HTTP router for private inter-service requests.
func (p *PersonApp) Private(store *store.Store) http.Handler {
	return http.NewServeMux()
}

// Log returns a logger configured to the preferences of the app.
func (p *PersonApp) Log() *slog.Logger {
	return p.log
}

// insertHandler is the request handler for insertPerson
func (p *PersonApp) insertHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rdm := rand.Intn(9999)

		id, err := store.Exec("InsertPerson", "Rick", "Sanchez", fmt.Sprintf("rick%s@sanchez.com", strconv.Itoa(rdm)))
		if err != nil {
			p.log.Error(errors.Wrap(err, "failed to Exec InsertPerson").Error())
			w.WriteHeader(http.StatusInternalServerError)
		}

		w.Write([]byte(fmt.Sprintf("Inserted record with ID %d", id)))
	}
}

func (p *PersonApp) getHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")

		prs, err := store.Exec("GetPerson", id)
		if err != nil {
			p.log.Error(errors.Wrap(err, "failed to Exec GetPerson").Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		person := prs.(*Person)

		resp, err := json.Marshal(person)
		if err != nil {
			p.log.Error(errors.Wrap(err, "failed to json.Marshal").Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}

func (p *PersonApp) selectHandler(store *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ppl, err := store.Exec("SelectPeople")
		if err != nil {
			p.log.Error(errors.Wrap(err, "failed to Exec GetPerson").Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		people := ppl.([]Person)

		resp, err := json.Marshal(people)
		if err != nil {
			p.log.Error(errors.Wrap(err, "failed to json.Marshal").Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(resp)
	}
}
