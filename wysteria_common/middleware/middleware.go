package middleware

import (
	"fmt"
	"errors"
	"time"
)

var (
	Timeout = time.Second * 30
	connectors = map[string] func(MiddlewareSettings) (WysteriaMiddleware, error) {
		"nats": natsConnect,
	}
)

func Connect(settings MiddlewareSettings) (WysteriaMiddleware, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}

	mid, err := connector(settings)
	if settings.EncryptionKey != "" {
		err = mid.SetKey(settings.EncryptionKey)
		if err != nil {
			return nil, err
		}
	}

	return mid, err
}

type WysteriaMiddleware interface {
	// Kill the connection to the middleware
	Close() error

	// Set encryption key for messages (if not set, encryption not used)
	SetKey(string) error

	// Send message, dont wait for reply
	Publish(subject string, data []byte) error

	// Subscribe to chan, each message received by only one listener
	Subscribe(subject, queue string) (WysteriaSubscription, error)

	// Subscribe to chan, each message received by all listeners
	GroupSubscribe(subject string) (WysteriaSubscription, error)

	// Send a message, subscribe to chan and for reply
	Request(subject string, data []byte) ([]byte, error)
}

type WysteriaSubscription interface {
	Unsubscribe() error
	Receive() <-chan Message
}
