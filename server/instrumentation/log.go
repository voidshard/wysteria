package instrumentation

import (
	"log"
	"os"
	"path/filepath"
)

const logdivider = "|"

// The most trivial of implementations - write via vanilla log package.
//
type simpleLogger struct {
	logger     *log.Logger
	filehandle *os.File
}

// Check if a dir exists. If not, create it with the given perms.
//
func ensureDirectory(path string, mode os.FileMode) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create the dir
		err = os.Mkdir(path, mode)
		if err != nil {
			return err
		}

		// Chmod to work around potential umask issues
		return os.Chmod(path, mode)
	}
	return nil
}

// Open a file if it exists, possibly create it first.
//
func openFile(path string) (*os.File, error) {
	var fh *os.File

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		// Create the log file
		tmp, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		fh = tmp
	} else {
		tmp, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, err
		}
		fh = tmp
	}

	return fh, nil
}

// Create a new simple logger given a log.Logger to write to
//
func newFileLogger(settings *Settings) (MonitorOutput, error) {
	err := ensureDirectory(settings.Location, 0775)
	if err != nil {
		return nil, err
	}

	fh, err := openFile(filepath.Join(settings.Location, settings.Target))
	if err != nil {
		return nil, err
	}

	return &simpleLogger{
		logger:     log.New(fh, "", log.LstdFlags),
		filehandle: fh,
	}, nil
}

// Log event to file.
//
func (l *simpleLogger) Log(doc *event) {
	l.logger.Println(logdivider,
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
func (l *simpleLogger) Close() {
	l.filehandle.Close()
}
