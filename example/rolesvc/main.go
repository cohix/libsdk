package main

import (
	"log"
	"log/slog"

	"github.com/cohix/libsdk/pkg/service"
	"github.com/pkg/errors"
)

func main() {
	svc, err := service.New("ROLES")
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to service.New"))
	}

	app := &RolesApp{
		log: slog.With("app", "ROLES"),
	}

	slog.Info("starting ROLES service")

	if err := svc.Serve(app); err != nil {
		log.Fatal(errors.Wrap(err, "failed to svc.Serve"))
	}
}
