package instrumentation

import (
	"context"
	"errors"
	"fmt"
	"github.com/olivere/elastic"
)

const defaultMapping = `{
	"mappings": {
		"wysterialog": {
			"properties": {
				"EpochTime": {"type": "date", "format": "epoch_millis"},
				"CallTarget": {"type": "string"},
				"CallType": {"type": "string"},
				"TimeTaken": {"type": "integer"},
				"InFunc": {"type": "string"},
				"Msg": {"type": "string"},
				"Severity": {"type": "string"},
				"Note": {"type": "string"},
				"UTCTime": {"type": "string"}
			}
		}
	}
}`

// Write log document to Elastic
//
type elasticLogger struct {
	url    string
	index  string
	client *elastic.Client
}

// Create a client to write data to elastic
//
func newElasticLogger(settings *Settings) (MonitorOutput, error) {
	client, err := elastic.NewClient(
		// Set elastic url to the given host / port
		elastic.SetURL(settings.Location),

		// If enabled there is some race condition in the underlying lib that can cause use to always
		// attempt to connect to the localhost. Assuming you're not running ES on the localhost you'll be
		// given a "no available hosts" error or something.
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}

	e := &elasticLogger{
		url:    settings.Location,
		index:  settings.Target,
		client: client,
	}
	return e, e.createIndexIfNotExists(settings.Target)
}

// Checks if an index with the given name exists. If not, creates the index.
func (l *elasticLogger) createIndexIfNotExists(index string) error {
	ctx := context.Background()

	exists, err := l.client.IndexExists(index).Do(ctx)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	createIndex, err := l.client.CreateIndex(index).BodyString(defaultMapping).Do(ctx)
	if err != nil {
		return err
	}
	if createIndex.Acknowledged {
		return nil
	}

	return errors.New(fmt.Sprintf("Creation of index %s not acknowledged", index))
}

func (l *elasticLogger) Log(doc *event) {
	l.client.Index().Index(l.index).Type(l.index).BodyJson(doc).Do(context.Background())
}

func (l *elasticLogger) Close() {
	l.client.Stop()
}
