package middleware

import (
	"crypto/tls"
	"fmt"
	"github.com/nats-io/nats"
	"log"
	"strconv"
)

const (
	url_prefix = "nats"
)

type natsMiddleware struct {
	Host    string
	key     *[32]byte
	conn    *nats.Conn
	encrypt bool
}

func natsUrl(settings *MiddlewareSettings) string {
	// eg nats://derek:pass@localhost:4222 (from nats docs)
	url := fmt.Sprintf("%s://", url_prefix)
	if settings.User != "" {
		url += settings.User
		if settings.Pass != "" {
			url += ":" + settings.Pass
		}
		url += "@"
	}
	return url + settings.Host + ":" + strconv.Itoa(settings.Port)
}

func natsConnect(settings *MiddlewareSettings) (WysteriaMiddleware, error) {
	options := []nats.Option{}
	if settings.PemFile != "" {
		options = append(options, nats.Secure(&tls.Config{
			InsecureSkipVerify: true,
		}))
	}

	raw, err := nats.Connect(
		natsUrl(settings),
		options...,
	)
	if err != nil {
		return nil, err
	}
	return &natsMiddleware{Host: settings.Host, conn: raw}, nil
}

func (m *natsMiddleware) Close() error {
	m.conn.Flush()
	m.conn.Close()
	return nil
}

func (m *natsMiddleware) Publish(subject string, data []byte) error {
	if m.encrypt {
		encoded, err := encrypt(data, m.key)
		if err != nil {
			return err
		}
		return m.conn.Publish(subject, encoded)
	}
	return m.conn.Publish(subject, data)
}

func (m *natsMiddleware) Request(subject string, data []byte) ([]byte, error) {
	if m.encrypt {
		encoded, err := encrypt(data, m.key)
		if err != nil {
			return []byte{}, err
		}

		msg, err := m.conn.Request(subject, encoded, Timeout)
		if err != nil {
			return []byte{}, err
		}

		clear, err := decrypt(msg.Data, m.key)
		return clear, nil
	}
	msg, err := m.conn.Request(subject, data, Timeout)
	return msg.Data, err
}

func (m *natsMiddleware) SetKey(key string) error {
	bkey, err := formKey(key)
	if err != nil {
		return err
	}
	m.encrypt = true
	m.key = &bkey
	return nil
}

// One subscriber to subject receives message
func (m *natsMiddleware) Subscribe(subject, queue string) (WysteriaSubscription, error) {
	// make an inbound chan
	recv := make(chan Message)
	f := func(msg *nats.Msg) {
		if m.encrypt {
			plain, err := decrypt(msg.Data, m.key)
			if err == nil {
				recv <- Message{Data: plain, Subject: msg.Subject, Reply: msg.Reply}
				return
			}
			log.Printf("[failure to decrypt] Subject: %s  Reply: %s", msg.Subject, msg.Reply)
		} else {
			recv <- Message{Data: msg.Data, Subject: msg.Subject, Reply: msg.Reply}
		}
	}

	subRecv, err := m.conn.QueueSubscribe(subject, queue, f)
	if err != nil {
		return nil, err
	}

	msgSub := natsSubscription{
		parent:  m,
		sub:     subRecv,
		Out:     recv,
		Subject: subject,
		Queue:   queue,
	}
	return &msgSub, nil
}

// All subscribers to subject receive all messages
func (m *natsMiddleware) GroupSubscribe(subject string) (WysteriaSubscription, error) {
	// make an inbound chan
	recv := make(chan Message)
	subRecv, err := m.conn.Subscribe(subject, func(msg *nats.Msg) {
		if m.encrypt {
			plain, err := decrypt(msg.Data, m.key)
			if err == nil {
				recv <- Message{Data: plain, Subject: msg.Subject, Reply: msg.Reply}
				return
			}
			log.Printf("[failure to decrypt] Subject: %s  Reply: %s", msg.Subject, msg.Reply)
		} else {
			recv <- Message{Data: msg.Data, Subject: msg.Subject, Reply: msg.Reply}
		}
	})
	if err != nil {
		return nil, err
	}

	msgSub := natsSubscription{
		parent:  m,
		sub:     subRecv,
		Out:     recv,
		Subject: subject,
	}
	return &msgSub, nil
}

type natsSubscription struct {
	parent  *natsMiddleware
	sub     *nats.Subscription
	Subject string
	Queue   string
	Out     <-chan Message
}

func (m *natsSubscription) Unsubscribe() error {
	return m.sub.Unsubscribe()
}

func (m *natsSubscription) Receive() <-chan Message {
	return m.Out
}
