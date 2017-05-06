package searchends

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
	"gopkg.in/olivere/elastic.v2"
	"strings"
	"sync"
	"time"
)

const (
	errDeleteBackoff = time.Second * 5
)

func elasticConnect(settings *Settings) (Searchbase, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(fmt.Sprintf("http://%s:%d", settings.Host, settings.Port)),
		elastic.SetSniff(false),
	)
	if err != nil {
		return nil, err
	}

	e := &elasticSearch{
		Settings: settings,
		client:   client,
	}

	err = e.createIndexIfNotExists(e.Settings.Database)
	if err != nil {
		return nil, err
	}

	return e, nil
}

type elasticSearch struct {
	Settings *Settings
	client   *elastic.Client
}

func (e *elasticSearch) InsertCollection(id string, doc *wyc.Collection) error {
	doc.Name = b64encode(doc.Name)
	return e.insert(tableCollection, id, doc)
}

func (e *elasticSearch) InsertItem(id string, doc *wyc.Item) error {
	doc.ItemType = b64encode(doc.ItemType)
	doc.Variant = b64encode(doc.Variant)
	encoded_facets := map[string]string{}
	for k, v := range doc.Facets {
		encoded_facets[b64encode(k)] = b64encode(v)
	}
	doc.Facets = encoded_facets
	return e.insert(tableItem, id, doc)
}

func (e *elasticSearch) InsertVersion(id string, doc *wyc.Version) error {
	encoded_facets := map[string]string{}
	for k, v := range doc.Facets {
		encoded_facets[b64encode(k)] = b64encode(v)
	}
	doc.Facets = encoded_facets
	return e.insert(tableVersion, id, doc)
}

func (e *elasticSearch) InsertResource(id string, doc *wyc.Resource) error {
	doc.Name = b64encode(doc.Name)
	doc.ResourceType = b64encode(doc.ResourceType)
	doc.Location = b64encode(doc.Location)
	return e.insert(tableFileresource, id, doc)
}

func (e *elasticSearch) InsertLink(id string, doc *wyc.Link) error {
	doc.Name = b64encode(doc.Name)
	return e.insert(tableLink, id, doc)
}

func (e *elasticSearch) UpdateItem(id string, doc *wyc.Item) error {
	// Explicit insert to ID deletes & replaces doc
	return e.insert(tableItem, id, doc)
}

func (e *elasticSearch) UpdateVersion(id string, doc *wyc.Version) error {
	// Explicit insert to ID deletes & replaces doc
	return e.insert(tableVersion, id, doc)
}

func (e *elasticSearch) DeleteCollection(ids ...string) error {
	return e.generic_delete(tableCollection, ids...)
}

func (e *elasticSearch) DeleteItem(ids ...string) error {
	return e.generic_delete(tableItem, ids...)
}

func (e *elasticSearch) DeleteVersion(ids ...string) error {
	return e.generic_delete(tableVersion, ids...)
}

func (e *elasticSearch) DeleteResource(ids ...string) error {
	return e.generic_delete(tableFileresource, ids...)
}

func (e *elasticSearch) DeleteLink(ids ...string) error {
	return e.generic_delete(tableLink, ids...)
}

func (e *elasticSearch) QueryCollection(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(tableCollection, elasticTermsCollection, limit, from, qs...)
}

func (e *elasticSearch) QueryItem(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(tableItem, elasticTermsItem, limit, from, qs...)
}

func (e *elasticSearch) QueryVersion(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(tableVersion, elasticTermsVersion, limit, from, qs...)
}

func (e *elasticSearch) QueryResource(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(tableFileresource, elasticTermsResource, limit, from, qs...)
}

func (e *elasticSearch) QueryLink(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(tableLink, elasticTermsLink, limit, from, qs...)
}

func (e *elasticSearch) Close() error {
	e.client.Stop()
	return nil
}

// Delete the given IDs.
//  - If the delete for an ID fails possibly it wasn't found as Elastic is still indexing it
//  - To overcome this we'll retry the delete once after a small sleep period
//
func (e *elasticSearch) generic_delete(col string, ids ...string) error {
	wg := sync.WaitGroup{}
	wg.Add(len(ids))

	err_chan := make(chan error)
	all_errors := []error{}

	go func() {
		for err := range err_chan {
			fmt.Println(err)
			all_errors = append(all_errors, err)
		}
	}()

	for _, id := range ids {
		my_id := id
		go func() {
			_, err := e.client.Delete().Index(e.Settings.Database).Type(col).Id(my_id).Do()
			if err != nil {
				// Delete called too quickly? Give elastic time to index & retry
				time.Sleep(errDeleteBackoff)
				_, err = e.client.Delete().Index(e.Settings.Database).Type(col).Id(my_id).Do()
			}
			if err != nil {
				err_chan <- err
			}
			wg.Done()
		}()
	}

	wg.Wait()
	close(err_chan)

	if len(all_errors) > 0 {
		return all_errors[0]
	}
	return nil
}

func (e *elasticSearch) insert(col, id string, doc interface{}) error {
	_, err := e.client.Index().Index(e.Settings.Database).Type(col).BodyJson(doc).Id(id).Do()
	return err
}

func (e *elasticSearch) createIndexIfNotExists(index string) error {
	exists, err := e.client.IndexExists(index).Do()
	if err != nil {
		return err
	}
	if exists {
		return nil
	}

	createIndex, err := e.client.CreateIndex(index).Do()
	if err != nil {
		return err
	}
	if createIndex.Acknowledged {
		return nil
	}

	return errors.New(fmt.Sprintf("Creation of index %s not acknowledged", index))
}

func elasticTermsResource(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.Name != "" {
		q = append(q, termQuery("Name", b64encode(qd.Name)))
	}
	if qd.ResourceType != "" {
		q = append(q, termQuery("ResourceType", b64encode(qd.ResourceType)))
	}
	if qd.Parent != "" {
		q = append(q, termQuery("Parent", qd.Parent))
	}
	if qd.Location != "" {
		q = append(q, termQuery("Location", b64encode(qd.Location)))
	}

	return q
}

func elasticTermsLink(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.Name != "" {
		q = append(q, termQuery("Name", b64encode(qd.Name)))
	}
	if qd.LinkSrc != "" {
		q = append(q, termQuery("Src", qd.LinkSrc))
	}
	if qd.LinkDst != "" {
		q = append(q, termQuery("Dst", qd.LinkDst))
	}
	return q
}

func elasticTermsCollection(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.Name != "" {
		q = append(q, termQuery("Name", b64encode(qd.Name)))
	}
	return q
}

func elasticTermsItem(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.ItemType != "" {
		q = append(q, termQuery("ItemType", b64encode(qd.ItemType)))
	}
	if qd.Variant != "" {
		q = append(q, termQuery("Variant", b64encode(qd.Variant)))
	}
	for k, v := range qd.Facets {
		key_encoded := b64encode(k)
		val_encoded := b64encode(v)
		q = append(q, termQuery(fmt.Sprintf("Facets.%s", key_encoded), val_encoded))
	}
	if qd.Parent != "" {
		q = append(q, termQuery("Parent", qd.Parent))
	}
	return q
}

func elasticTermsVersion(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.VersionNumber > 0 {
		q = append(q, elastic.NewTermQuery("Number", qd.VersionNumber))
	}
	for k, v := range qd.Facets {
		key_encoded := b64encode(k)
		val_encoded := b64encode(v)
		q = append(q, termQuery(fmt.Sprintf("Facets.%s", key_encoded), val_encoded))
	}
	if qd.Parent != "" {
		q = append(q, termQuery("Parent", qd.Parent))
	}
	return q
}

func termQuery(k, v string) elastic.TermQuery {
	// Elastic will have tokenized the string(s) so we'll lowercase our string to search for
	return elastic.NewTermQuery(k, strings.ToLower(v))
}

// Send query to ElasticSearch
//  - We only ever return IDs (our search db isn't our canonical data source)
//  - Check to ensure we don't return duplicate IDs
//  - The terms of each wyc.QueryDesc are joined via Bool query "MUST" thus are "AND" together
//  - Because we concatenate all results, multiple wyc.QueryDesc form an "OR"
//
func (e *elasticSearch) fanSearch(table string, makeTerms func(*wyc.QueryDesc) []elastic.TermQuery, limit int, from int, queries ...*wyc.QueryDesc) ([]string, error) {
	// Our query base - specify the index, page, limits & desired fields
	base := e.client.Search().Index(e.Settings.Database).Type(table).Fields("Id").From(from)
	if limit > 0 {
		base.Size(limit)
	}

	// Ultimately we're doing an "or" query where we're after any result that matches all
	// the fields of at least one of our QueryDesc
	or_query := elastic.NewBoolQuery()
	for _, query := range queries {

		// Represents an individual QueryDesc getting made into a "must" term query
		bquery := elastic.NewBoolQuery()
		for _, q := range makeTerms(query) {
			bquery = bquery.Must(q)
		}

		or_query = or_query.Should(bquery)
	}

	results := []string{}

	// Finally, perform the query
	res, err := base.Query(or_query).Do()
	if err != nil {
		return results, err
	} else {
		// And pull together all the results
		for _, hit := range res.Hits.Hits {
			// ToDo: works, but could use some straightening up
			// (We only asked for the Id field)
			result_id := hit.Fields["Id"].([]interface{})[0].(string)
			results = append(results, result_id)
		}
	}

	return results, nil
}
