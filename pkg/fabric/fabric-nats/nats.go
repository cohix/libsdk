package fabricnats

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/cohix/libsdk/pkg/fabric"
	"github.com/gofrs/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/pkg/errors"
)

// local is the default (running on the same system)
const localNatsAddr = nats.DefaultURL
const natsAddrEnvKey = "LIBSDK_FABRIC_NATS_ADDR"

var _ fabric.Fabric = &Nats{}

type Nats struct {
	serviceName string
	nc          *nats.Conn
	js          jetstream.JetStream
	s           jetstream.Stream
}

type MsgConnection struct{}

// ReplayConnection is a connection for pub/sub/replay
type ReplayConnection struct {
	log      slog.Logger
	subject  string
	stream   jetstream.Stream
	consumer jetstream.Consumer
	info     *jetstream.ConsumerInfo
	publish  func(ctx context.Context, subject string, payload []byte, opts ...jetstream.PublishOpt) (*jetstream.PubAck, error)
}

// New creates a new NATS fabric
func New(serviceName string) (*Nats, error) {
	natsAddr := localNatsAddr
	if envAddr, exists := os.LookupEnv(natsAddrEnvKey); exists {
		natsAddr = envAddr
	}

	nc, err := nats.Connect(natsAddr)
	if err != nil {
		return nil, errors.Wrap(err, "failed to nats.Connect")
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, errors.Wrap(err, "failed to jetstream.New")
	}

	s, err := js.CreateOrUpdateStream(context.Background(), jetstream.StreamConfig{
		Name: serviceName,
		// Subjects for SERVICE.store, and SERVICE.pub are attached to preserve their state. SERVICE.msg subjects are not persisted.
		Subjects:    []string{fmt.Sprintf("%s.store", serviceName), fmt.Sprintf("%s.pub", serviceName)},
		Storage:     jetstream.FileStorage,
		Retention:   jetstream.LimitsPolicy,
		MaxBytes:    32000000000, // 32GB
		Compression: jetstream.S2Compression,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to CreateOrUpdateStream")
	}

	n := &Nats{
		serviceName: serviceName,
		nc:          nc,
		js:          js,
		s:           s,
	}

	return n, nil
}

// Messenger returns a connection for message sending/receiving
func (n *Nats) Messenger(service string) (fabric.MsgConnection, error) {
	return &MsgConnection{}, nil
}

// Replayer returns a connection for Replayer publish/replay
func (n *Nats) Replayer(subject string, beginning bool) (fabric.ReplayConnection, error) {
	ctx := context.Background()

	// all consumers are unique, even if there are multiple within
	// a single server instance. For example, store and pub consumers
	// have different behaviour and therefore have unique names
	consumerUUID, err := uuid.NewV7()
	if err != nil {
		return nil, errors.Wrap(err, "failed to uuid.NewV7")
	}

	consumerName := fmt.Sprintf("%s-consumer-%s", n.serviceName, consumerUUID.String())
	fullSubject := fmt.Sprintf("%s.%s", n.serviceName, subject)

	deliverPolicy := jetstream.DeliverAllPolicy

	if !beginning {
		deliverPolicy = jetstream.DeliverNewPolicy
	}

	c, err := n.s.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       consumerName,
		DeliverPolicy: deliverPolicy,
		AckPolicy:     jetstream.AckExplicitPolicy,
		FilterSubject: fullSubject,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to OrderedConsumer")
	}

	// get consumer info to power the upTo channel in the replayer
	info, err := c.Info(context.Background())
	if err != nil {
		return nil, errors.Wrap(err, "failed to consumer.Info")
	}

	b := &ReplayConnection{
		log:      *slog.With("lib", "libsdk", "module", "fabricnats"),
		subject:  fullSubject,
		stream:   n.s,
		consumer: c,
		info:     info,
		publish:  n.js.Publish,
	}

	return b, nil
}

func (m *MsgConnection) SendAndRecv(msg any, receiver fabric.Receiver) error {
	return nil
}

func (m *MsgConnection) RecvAndReply(handler fabric.Handler) {

}

// Publish publishes a message to a broadcast channel
func (b *ReplayConnection) Publish(msg any) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return errors.Wrap(err, "failed to json.Marshal")
	}

	_, err = b.publish(context.Background(), b.subject, body)
	if err != nil {
		return errors.Wrap(err, "failed to publish")
	}

	return nil
}

func (b *ReplayConnection) Replay(gen fabric.Generator, recv fabric.Receiver) (chan bool, error) {
	upToChan := make(chan bool, 1)
	upToOnce := sync.Once{}
	upToCounter := uint64(0)

	upToCompletion := func() {
		upToOnce.Do(func() {
			upToChan <- true
		})
	}

	// in the special case where this is the very first time a service is
	// running and there are no messages in the stream at all, notify now
	if b.info.NumPending == 0 {
		upToCompletion()
	}

	msgs, err := b.consumer.Messages()
	if err != nil {
		return nil, errors.Wrap(err, "failed to consumer.Messages")
	}

	go func() {
		for {
			msg, err := msgs.Next()
			if err != nil {
				b.log.Error(errors.Wrap(err, "failed to msgs.Next").Error())
				continue
			}

			upToCounter++
			msg.Ack()

			// notify the caller when we've reached the point of the
			// stream where we attached to it as a new consumer
			// but only once as we'd be blocking message reading otherwise
			if upToCounter >= b.info.NumPending {
				upToCompletion()
			}

			// grab a typed object from the replay consumer
			// via the generator into which we unmarshal the data
			obj := gen()

			if err := json.Unmarshal(msg.Data(), obj); err != nil {
				b.log.Error(errors.Wrap(err, "failed to json.Unmarshal").Error())
				continue
			}

			recv(obj)
		}
	}()

	return upToChan, nil
}
