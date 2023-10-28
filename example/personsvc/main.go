package main

import (
	"fmt"
	"log"
	"log/slog"
	"math/rand"
	"strconv"

	"github.com/cohix/libsdk/example"
	"github.com/cohix/libsdk/pkg/fabric"
	fabricnats "github.com/cohix/libsdk/pkg/fabric/fabric-nats"
	"github.com/cohix/libsdk/pkg/store"
	driversqlite "github.com/cohix/libsdk/pkg/store/driver-sqlite"
	"github.com/pkg/errors"
)

func main() {
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

	rdm := rand.Intn(9999)

	id, err := store.Exec("InsertPerson", "Rick", "Sanchez", fmt.Sprintf("rick%s@sanchez.com", strconv.Itoa(rdm)))
	if err != nil {
		log.Fatal(errors.Wrap(err, "failed to store.Exec"))
	}

	slog.Info("Inserted record with ID", "id", id)
}
