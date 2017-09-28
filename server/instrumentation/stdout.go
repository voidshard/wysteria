package instrumentation

import (
	"log"
)


func newStdoutLogger(s *Settings) (MonitorOutput, error) {
	return &stdoutLogger{}, nil
}

type stdoutLogger struct {}

// Log event to shell
//
func (l *stdoutLogger) Log(doc *event) {
	log.Println(logdivider,
		doc.EpochTime, logdivider,
		doc.TimeTaken, logdivider,
		doc.Severity, logdivider,
		doc.CallType, logdivider,
		doc.CallTarget, logdivider,
		doc.InFunc, logdivider,
		doc.Msg, logdivider,
		doc.Note,
	)
}

// Close open file
//
func (l *stdoutLogger) Close() {}
