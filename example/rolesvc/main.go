package main

import (
	"fmt"
	"log"
	"time"

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

	rpl, err := f.Replayer("store", true)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := rpl.Replay(msgGen, func(msg any) {
			realMsg := msg.(*message)

			fmt.Printf("Message from %s to %s: %s\n", realMsg.From, realMsg.To, realMsg.Text)
		}); err != nil {
			log.Fatal(err)
		}
	}()

	time.Sleep(time.Second * 10)
}

// generator for message handler
func msgGen() any {
	return &message{}
}
