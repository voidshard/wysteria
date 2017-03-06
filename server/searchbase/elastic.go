package searchends

import (
	"encoding/base64"
	"errors"
	"fmt"
	"gopkg.in/olivere/elastic.v2"
	"strings"
	"sync"
	"time"
	wyc "github.com/voidshard/wysteria/common"
)

const (
	err_delete_backoff = time.Second * 5
)

func elastic_connect(settings *SearchbaseSettings) (Searchbase, error) {
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
	Settings *SearchbaseSettings
	client   *elastic.Client
}

func (e *elasticSearch) InsertCollection(id string, doc *wyc.Collection) error {
	return e.insert(table_collection, id, doc)
}

func (e *elasticSearch) InsertItem(id string, doc *wyc.Item) error {
	return e.insert(table_item, id, doc)
}

func (e *elasticSearch) InsertVersion(id string, doc *wyc.Version) error {
	return e.insert(table_version, id, doc)
}

func (e *elasticSearch) InsertResource(id string, doc *wyc.Resource) error {
	// Hash path to nullify tokenizing on '/' or '\' symbols
	doc.Location = hashResourceLocation(doc.Location)
	return e.insert(table_fileresource, id, doc)
}

func (e *elasticSearch) InsertLink(id string, doc *wyc.Link) error {
	return e.insert(table_link, id, doc)
}

func (e *elasticSearch) UpdateItem(id string, doc *wyc.Item) error {
	// Explicit insert to ID deletes & replaces doc
	return e.insert(table_item, id, doc)
}

func (e *elasticSearch) UpdateVersion(id string, doc *wyc.Version) error {
	// Explicit insert to ID deletes & replaces doc
	return e.insert(table_version, id, doc)
}

func (e *elasticSearch) DeleteCollection(ids ...string) error {
	return e.delete(table_collection, ids...)
}

func (e *elasticSearch) DeleteItem(ids ...string) error {
	return e.delete(table_item, ids...)
}

func (e *elasticSearch) DeleteVersion(ids ...string) error {
	return e.delete(table_version, ids...)
}

func (e *elasticSearch) DeleteResource(ids ...string) error {
	return e.delete(table_fileresource, ids...)
}

func (e *elasticSearch) DeleteLink(ids ...string) error {
	return e.delete(table_link, ids...)
}

func (e *elasticSearch) QueryCollection(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(table_collection, elasticTermsCollection, limit, from, qs...)
}

func (e *elasticSearch) QueryItem(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(table_item, elasticTermsItem, limit, from, qs...)
}

func (e *elasticSearch) QueryVersion(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(table_version, elasticTermsVersion, limit, from, qs...)
}

func (e *elasticSearch) QueryResource(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(table_fileresource, elasticTermsResource, limit, from, qs...)
}

func (e *elasticSearch) QueryLink(limit, from int, qs ...*wyc.QueryDesc) ([]string, error) {
	return e.fanSearch(table_link, elasticTermsLink, limit, from, qs...)
}

func (e *elasticSearch) Close() error {
	e.client.Stop()
	return nil
}

// Delete the given IDs.
//  - If the delete for an ID fails possibly it wasn't found as Elastic is still indexing it
//  - To overcome this we'll retry the delete once after a small sleep period
//
func (e *elasticSearch) delete(col string, ids ...string) error {
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
				time.Sleep(err_delete_backoff)
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

func hashResourceLocation(path string) string {
	// avoid any odd chars or tokenizing by Elastic by b64 encoding the string
	return strings.TrimRight(base64.StdEncoding.EncodeToString([]byte(path)), "=")
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
		q = append(q, termQuery("Name", qd.Name))
	}
	if qd.ResourceType != "" {
		q = append(q, termQuery("ResourceType", qd.ResourceType))
	}
	if qd.Parent != "" {
		q = append(q, termQuery("Parent", qd.Parent))
	}
	if qd.Location != "" {
		hsh := hashResourceLocation(qd.Location)
		q = append(q, termQuery("Location", hsh))
	}

	return q
}

func elasticTermsLink(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.Name != "" {
		q = append(q, termQuery("Name", qd.Name))
	}
	if qd.LinkSrc != "" {
		q = append(q, termQuery("Src", qd.LinkSrc))
	}
	if qd.LinkDst != "" {
		q = append(q, termQuery("Dest", qd.LinkDst))
	}
	return q
}

func elasticTermsCollection(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.Name != "" {
		q = append(q, termQuery("Name", qd.Name))
	}
	return q
}

func elasticTermsItem(qd *wyc.QueryDesc) (q []elastic.TermQuery) {
	if qd.Id != "" {
		q = append(q, elastic.NewTermQuery("Id", qd.Id))
	}
	if qd.ItemType != "" {
		q = append(q, termQuery("ItemType", qd.ItemType))
	}
	if qd.Variant != "" {
		q = append(q, termQuery("Variant", qd.Variant))
	}
	for k, v := range qd.Facets {
		q = append(q, termQuery(fmt.Sprintf("Facets.%s", k), v))
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
		q = append(q, termQuery(fmt.Sprintf("Facets.%s", k), v))
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

// Fan out elastic search,
//  - Execute all queries from each wyc.QueryDesc in parallel & concatenate results
//  - We only ever return IDs (our search db isn't our canonical data source)
//  - Check to ensure we don't return duplicate IDs
//  - ToDo: At the moment we only return the first err (if there are any)
//  - The terms of each wyc.QueryDesc are joined via Bool query "MUST" thus are "AND" together
//  - Because we concatenate all results, multiple wyc.QueryDesc form an "OR"
//  - ToDo: Possibly Elastic could return the answer set faster with a more elaborate query
//
func (e *elasticSearch) fanSearch(table string, makeTerms func(*wyc.QueryDesc) []elastic.TermQuery, limit int, from int, queries ...*wyc.QueryDesc) ([]string, error) {
	result_chan := make(chan *elastic.SearchResult, len(queries))
	err_chan := make(chan error)
	rwg := sync.WaitGroup{} // wait group to ensure we've finished compiling results
	rwg.Add(1)

	wg := sync.WaitGroup{} // wait group to ensure we've completed all queries
	wg.Add(len(queries))

	all_errors := []error{}   // list of errors returned
	results := []string{} // list of results (Ids) returned

	go func() {
		// Listen for errors, record them all
		for err := range err_chan {
			all_errors = append(all_errors, err)
		}
	}()

	go func() {
		// Listen for results, pull out the IDs and record those we don't have already
		for result := range result_chan {
			for _, hit := range result.Hits.Hits {
				// ToDo: works, but could use some straightening up
				result_id := hit.Fields["Id"].([]interface{})[0].(string)

				add := true
				for _, id := range results {
					if id == result_id {
						add = false
						break
					}
				}
				if add {
					results = append(results, result_id)
				}
			}
		}
		rwg.Done()
	}()

	for _, query := range queries {
		// Here we go through our distinct queries to execute and run them all in their own routines.
		// Loop through our queries
		//  - build a list of QueryTerms w/ makeTerms func
		//  - pass query terms to BoolQuery (w/ Must)
		//  - execute the query
		//  - if there was an error, pass to the err chan
		//  - if there are result(s), pass them to the result chan
		s_query := query
		go func() {
			s := e.client.Search().Index(e.Settings.Database).Type(table).Fields("Id")
			
			if limit > 0 {
				s = s.Size(limit)
			}
			s = s.From(from)
			
			bquery := elastic.NewBoolQuery()
			for _, q := range makeTerms(s_query) {
				bquery.Must(q)
			}

			res, err := s.Query(bquery).Do()
			if err != nil {
				err_chan <- err
			} else {
				if len(res.Hits.Hits) > 0 {
					result_chan <- res
				}
			}
			wg.Done()
		}()
	}

	wg.Wait() // wait for all queries to come back

	close(result_chan)
	close(err_chan)

	rwg.Wait() // wait for all results to be parsed

	if len(all_errors) > 0 {
		return results, all_errors[0]
	}
	return results, nil
}
