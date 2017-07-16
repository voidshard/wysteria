package instrumentation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
)

const (
	// We'll only hold this many docs in memory (they're still written to logs ASAP)
	MonitorMaxDocsHeld = 50

	// Severity constants
	severityError = "error"
	severityWarn  = "warning"
	severityInfo  = "info"
)

// Config for monitor's webserver
//
type WebserverConfig struct {
	Port           int
	EndpointHealth string
}

// Parent struct that handles writing to all child MonitorOutput(s)
//
type Monitor struct {
	incoming chan *event

	lock    sync.Mutex
	outputs []MonitorOutput // all available outputs
	recent  []*event        // most recent docs
	maxLen  int             // max number of docs held internally
	latest  int             // last doc saved internally
}

// Create and return a new Monitor
//  Kicks off child routine to write new log messages
//
func NewMonitor(outs ...MonitorOutput) (*Monitor, error) {
	if len(outs) < 1 {
		return nil, errors.New("Require at least one MonitorOutput to be configured")
	}

	m := &Monitor{
		lock:     sync.Mutex{},
		outputs:  outs,
		maxLen:   MonitorMaxDocsHeld,
		recent:   make([]*event, MonitorMaxDocsHeld),
		incoming: make(chan *event),
	}
	return m, nil
}

// Pull down the Monitor & all outputs
//
func (m *Monitor) Stop() {
	m.lock.Lock()
	defer m.lock.Unlock()

	close(m.incoming)
	for _, out := range m.outputs {
		out.Close()
	}
}

// Build health / stats reply
//
func (m *Monitor) prepareReport() ([]byte, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	logs := []*event{}
	for i := 0; i < m.maxLen; i++ {
		index := (i + m.latest) % m.maxLen
		if m.recent[index] == nil {
			break
		}
		logs = append(logs, m.recent[index])
	}

	return json.Marshal(map[string]interface{}{
		"logs":   logs,
		"status": "OK",
	})
}

//
//
func (m *Monitor) healthCheck(w http.ResponseWriter, r *http.Request) {
	logdata, err := m.prepareReport()
	if err != nil {
		w.WriteHeader(500)
		io.WriteString(w, err.Error())
		return
	}

	w.WriteHeader(200)
	w.Write(logdata)
}

// Kick off the sub routines that run int the Monitor
//  - routines to write to log(s)
//  - routine to reply to http requests
//
func (m *Monitor) Start(settings *WebserverConfig) error {
	go m.run() // handle incoming log messages
	go m.run() // handle incoming log messages

	go func() { // serve up http endpoints
		http.HandleFunc(settings.EndpointHealth, m.healthCheck)
		err := http.ListenAndServe(fmt.Sprintf(":%d", settings.Port), nil)
		if err != nil {
			panic(err)
		}
	}()
	return nil
}

// Intended to be run via a single routine - which listens forever for all incoming
// docs and writes them to each available MonitorOutput mechanism.
//
func (m *Monitor) run() {
	for doc := range m.incoming {
		m.lock.Lock()
		m.latest += 1 // Considered Atomic num, but I think we'd have to lock to apply math on next line anyhow?
		m.latest %= m.maxLen
		m.recent[m.latest] = doc
		m.lock.Unlock()

		// ToDo: Do we need to lock here? I've tested this writing out 12+ logs w/ 20+ routines and it
		// doesn't seem to break for me ..
		for _, out := range m.outputs {
			out.Log(doc)
		}
	}
}

// Write given doc to all output(s)
// Since this is intended for logging, no errors due to being unable to log should cause failures.
//
func (m *Monitor) Log(msg string, opts ...Opt) {
	event := newEvent(msg)

	for _, o := range opts {
		o(event)
	}

	m.incoming <- event
}

// Log a warning message
//
func (m *Monitor) Warn(msg string, opts ...Opt) {
	event := newEvent(msg)

	for _, o := range opts {
		o(event)
	}

	event.Severity = severityWarn
	m.incoming <- event
}

// Log an error message
//
func (m *Monitor) Err(err error, opts ...Opt) {
	event := newEvent(err.Error())

	for _, o := range opts {
		o(event)
	}

	event.Severity = severityError
	m.incoming <- event
}
