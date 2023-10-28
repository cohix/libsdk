package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/cohix/libsdk/example"
	"github.com/cohix/libsdk/pkg/fabric"
	fabricnats "github.com/cohix/libsdk/pkg/fabric/fabric-nats"
	"github.com/cohix/libsdk/pkg/store"
	driversqlite "github.com/cohix/libsdk/pkg/store/driver-sqlite"
	"github.com/pkg/errors"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("args")
	}

	var f fabric.Fabric
	var err error

	f, err = fabricnats.New("SVC")
	if err != nil {
		log.Fatal(err)
	}

	rpl, err := f.Replayer("store", true)
	if err != nil {
		log.Fatal(err)
	}

	driver, err := driversqlite.New("SVC")
	if err != nil {
		log.Fatal(err)
	}

	store := store.New(driver, rpl)

	store.Register("InsertPerson", example.InsertPersonHandler)
	store.Register("GetPerson", example.GetPersonHandler)

	if err := store.Start(example.Migrations); err != nil {
		log.Fatal(errors.Wrap(err, "failed to store.Start"))
	}

	person, err := store.Exec("GetPerson", os.Args[1])
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to store.Exec"))
	}

	slog.Info("Got person", "person", person)
}
