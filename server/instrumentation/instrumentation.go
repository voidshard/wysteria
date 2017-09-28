/*
Instrumentation module
 - collects server stats on calls made, errors.
 - writes stats to various output(s).
 - keeps the most recent handful of documents around for viewing.
   That is, 'viewing' here refers to showing the most recent activity suitable for health checks via the HTTP
   health route. More advanced graphing of calls / viewing / Monitoring is left to more suitable systems
   (kibana, grafana .. etc) and isn't intended to be viewable directly via wysteria at any point.

Simply put a Monitor stores 'stats' documents via some MonitorOutput(s), and keeps the last
handful of documents handy for display via health checks.
*/

package instrumentation

import (
	"errors"
	"fmt"
	"time"
)

const (
	// The driver names that we accept
	DriverLogfile = "logfile"
	DriverElastic = "elastic"
	DriverStdout  = "stdout"

	// Call types
	callFind      = "find"
	callCreate    = "create"
	callDelete    = "delete"
	callPublish   = "publish"
	callPublished = "published"
	callUpdate    = "update"

	// What is being found / created / deleted / updated
	targetCollection = "collection"
	targetItem       = "item"
	targetVersion    = "version"
	targetResource   = "resource"
	targetLink       = "link"
)

var (
	// List of known drivers & the functions to start up a connection
	connectors = map[string]func(*Settings) (MonitorOutput, error){
		DriverLogfile: newFileLogger,
		DriverElastic: newElasticLogger,
		DriverStdout:  newStdoutLogger,
	}
)

// Return a connect function for the given settings, or err if it can't be found
//
func Connect(settings *Settings) (MonitorOutput, error) {
	connector, ok := connectors[settings.Driver]
	if !ok {
		return nil, errors.New(fmt.Sprintf("Connector not found for %s", settings.Driver))
	}
	return connector(settings)
}

type event struct {
	Msg        string
	Note       string
	InFunc     string
	Severity   string
	EpochTime  int64
	UTCTime    string
	TimeTaken  int64
	CallType   string
	CallTarget string
}

type Opt func(*event)

// Set InFunc value
//
func InFunc(msg string) Opt {
	return func(i *event) {
		i.InFunc = msg
	}
}

// Record the kind of function call we're making
//
func IsFind() Opt {
	return func(i *event) {
		i.CallType = callFind
	}
}

func IsCreate() Opt {
	return func(i *event) {
		i.CallType = callCreate
	}
}

func IsDelete() Opt {
	return func(i *event) {
		i.CallType = callDelete
	}
}

func IsUpdate() Opt {
	return func(i *event) {
		i.CallType = callUpdate
	}
}

func IsPublish() Opt {
	return func(i *event) {
		i.CallType = callPublish
	}
}

func IsPublished() Opt {
	return func(i *event) {
		i.CallType = callPublished
	}
}

func TargetCollection() Opt {
	return func(i *event) {
		i.CallTarget = targetCollection
	}
}

// Set what kind of obj we're targeting (trying to find, create, delete etc)
//
func TargetItem() Opt {
	return func(i *event) {
		i.CallTarget = targetItem
	}
}

func TargetVersion() Opt {
	return func(i *event) {
		i.CallTarget = targetVersion
	}
}

func TargetResource() Opt {
	return func(i *event) {
		i.CallTarget = targetResource
	}
}

func TargetLink() Opt {
	return func(i *event) {
		i.CallTarget = targetLink
	}
}

// Set Time (time taken in nano seconds) value
//
func Time(t int64) Opt {
	return func(i *event) {
		i.TimeTaken = t
	}
}

// Set Note value
//
func Note(msg string) Opt {
	return func(i *event) {
		i.Note = msg
	}
}

// Represents somewhere to write stats / log documents to
//
type MonitorOutput interface {
	Log(*event)
	Close()
}

// Create a new event with some default fields filled out
//
func newEvent(msg string) *event {
	tnow := time.Now()
	return &event{
		Msg:       msg,
		EpochTime: tnow.Unix() * 1000,
		UTCTime:   tnow.UTC().String(),
		Severity:  severityInfo,
	}
}

// Settings for some form of logging output
//
type Settings struct {
	// What will be used to write the log
	Driver string

	// What host / location the log should go to
	Location string

	// Where in the given location the log should go
	Target string
}
