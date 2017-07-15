package instrumentation

import (
	"gopkg.in/olivere/elastic.v2"
	"errors"
	"fmt"
)

// Write log document to Elastic
//
type elasticLogger struct {
	url string
	index string
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
		url: settings.Location,
		index: settings.Target,
		client: client,
	}
	return e, e.createIndexIfNotExists(settings.Target)
}

// Checks if an index with the given name exists. If not, creates the index.
func (l *elasticLogger) createIndexIfNotExists(index string) error {
	exists, err := l.client.IndexExists(index).Do()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	createIndex, err := l.client.CreateIndex(index).Do()
	if err != nil {
		return err
	}
	if createIndex.Acknowledged {
		return nil
	}

	return errors.New(fmt.Sprintf("Creation of index %s not acknowledged", index))
}

func (l *elasticLogger) Log(doc *event) {
	l.client.Index().Index(l.index).Type(l.index).BodyJson(doc).Do()
}

func (l *elasticLogger) Close() {
	l.client.Stop()
}
