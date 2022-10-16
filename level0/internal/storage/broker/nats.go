package broker

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

type Nats struct {
	EncodedConn *nats.EncodedConn
}

func (n *Nats) Close() {
	n.EncodedConn.Close()
}

// NewNatsConnect CONNECTING TO NATS-STREAMING-SERVER
func NewNatsConnect(natsConfig Config) (*Nats, error) {
	nc, err := nats.Connect(natsConfig.URL)
	if err != nil {
		return nil, fmt.Errorf("nats.Connect from main ERROR: %w", err)
	}

	sc, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		return nil, fmt.Errorf("encodedConnect from main ERROR: %w", err)
	}

	return &Nats{
		EncodedConn: sc,
	}, nil
}
