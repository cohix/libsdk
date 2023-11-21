package main

import (
	"log"
	"log/slog"

	"github.com/cohix/libsdk/pkg/service"
	"github.com/pkg/errors"
)

func main() {
	// creating a "service" also creates the underlying "fabric" and "store"
	// with the default NATS and SQLite drivers, respectively.
	svc, err := service.New("PERSON")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to service.New"))
	}

	// an app is a type that returns everything that a service needs to operate
	// such as migrations, registered transactions, public and private HTTP handlers, etc.
	app := &PersonApp{
		log: slog.With("app", "PERSON"),
	}

	slog.Info("starting PERSON service")

	// calling Serve causes a few things to happen:
	// 1) The fabric is started and waits for successful connection
	// 2) The store initializes, runs all the migrations provided by the app
	//	  and then plays back all of the historical transactions from the fabric
	// 3) Starts an HTTP server, handled by the app.Public() http.Handler
	if err := svc.Serve(app); err != nil {
		log.Fatal(errors.Wrap(err, "failed to svc.Serve"))
	}
}
