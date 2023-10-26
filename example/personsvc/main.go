package main

import (
	"log"
	"log/slog"
	"math/rand"
	"strconv"

	"github.com/cohix/libsdk/pkg/fabric"
	fabricnats "github.com/cohix/libsdk/pkg/fabric/fabric-nats"
)

type message struct {
	From string
	To   string
	Text string
}

func main() {
	var f fabric.Fabric
	var err error

	f, err = fabricnats.New("SVC")
	if err != nil {
		log.Fatal(err)
	}

	rpl, err := f.Replayer("store", false)
	if err != nil {
		log.Fatal(err)
	}

	rdm := rand.Intn(99999)

	if err := rpl.Publish(message{
		From: "Rick",
		To:   "Morty",
		Text: "hello " + strconv.Itoa(rdm),
	}); err != nil {
		log.Fatal(err)
	}

	slog.Info("published message!")
}
