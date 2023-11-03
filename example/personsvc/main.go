package main

import (
	"log"
	"log/slog"

	"github.com/cohix/libsdk/pkg/service"
	"github.com/pkg/errors"
)

func main() {
	svc, err := service.New("PERSON")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to service.New"))
	}

	app := &PersonApp{
		log: slog.With("app", "PERSON"),
	}

	slog.Info("starting PERSON service")

	if err := svc.Serve(app); err != nil {
		log.Fatal(errors.Wrap(err, "failed to svc.Serve"))
	}
}
