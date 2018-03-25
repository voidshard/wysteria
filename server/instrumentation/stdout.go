/*The most trivial of all loggers - writes event data to stdout

Produces lines like:
2018/03/18 17:46:20 | 1521395180000 | 4 | info | create | collection |  | database | [ 71415080-8d05-40c4-9615-adea42663b72 TestItemLinkTo1]
2018/03/18 17:46:20 | 1521395180000 | 4 | info | create | collection |  | searchbase | [ 71415080-8d05-40c4-9615-adea42663b72 TestItemLinkTo1]
2018/03/18 17:46:20 | 1521395180000 | 8 | info | create | collection |  | middleware | [TestItemLinkTo1]
2018/03/18 17:46:20 | 1521395180000 | 3 | info | create | item |  | database | [71415080-8d05-40c4-9615-adea42663b72 142316da-fd60-4ed1-a852-1aad732f3f43 super item]
2018/03/18 17:46:20 | 1521395180000 | 8 | info | create | item |  | searchbase | [71415080-8d05-40c4-9615-adea42663b72 142316da-fd60-4ed1-a852-1aad732f3f43 super item]
2018/03/18 17:46:20 | 1521395180000 | 12 | info | create | item |  | middleware | [71415080-8d05-40c4-9615-adea42663b72 142316da-fd60-4ed1-a852-1aad732f3f43 super item]
2018/03/18 17:46:20 | 1521395180000 | 3 | info | create | item |  | database | [71415080-8d05-40c4-9615-adea42663b72 6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15 house brick]
2018/03/18 17:46:20 | 1521395180000 | 8 | info | create | item |  | searchbase | [71415080-8d05-40c4-9615-adea42663b72 6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15 house brick]
2018/03/18 17:46:20 | 1521395180000 | 12 | info | create | item |  | middleware | [71415080-8d05-40c4-9615-adea42663b72 6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15 house brick]
2018/03/18 17:46:20 | 1521395180000 | 4 | info | create | link |  | database | [7177b4b2-17e0-4675-aa2f-da06ab91ac12 142316da-fd60-4ed1-a852-1aad732f3f43 6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15]
2018/03/18 17:46:20 | 1521395180000 | 9 | info | create | version |  | searchbase | [7177b4b2-17e0-4675-aa2f-da06ab91ac12 142316da-fd60-4ed1-a852-1aad732f3f43 6bf7a22e-0b8e-47c0-bc78-c4fb2219dd15]

Ie.
timestamp | epoch time | time taken in milliseconds | severity | action type | object being acted on | <place holder> | layer | generic info

Please note:
Usually time taken is logged in nano seconds, but to be more readable for a *human* the stdout logger converts this to millis.
The other loggers write this time taken in nano seconds, as commonly the millis return '0' .. which makes for bad graphs.
And as we all know, logging is all about the sweet sweet graphs.
*/

package instrumentation

import (
	"log"
	"time"
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
		doc.TimeTaken / int64(time.Millisecond), logdivider,
		doc.Severity, logdivider,
		doc.CallType, logdivider,
		doc.CallTarget, logdivider,
		doc.InFunc, logdivider,
		doc.Msg, logdivider,
		doc.toKeyValueString(), logdivider,
		doc.Note,
	)
}

// Close open file
//
func (l *stdoutLogger) Close() {}
