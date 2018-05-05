/*Use Elastic as a search backend. This is specifically for Elastic 6

Elastic includes the helpful notion of tokenizing everything we send it. This is a problem for us because we don't
really enforce any style of ID(s), Names, or other strings which elastic will happily token-ify based on symbols,
spaces or .. other such things. To get around this, we encode strings we send elastic with base64 to make them
continuous non-token-ify-able strings. Not the most efficient approach, but it does save on headaches.
*/

package searchends

import (
	"fmt"
	"context"
	wyc "github.com/voidshard/wysteria/common"
	"github.com/olivere/elastic"
)


func elasticSixConnect(settings *Settings) (Searchbase, error) {
	clt, err := elastic.NewClient(
		// Set elastic url to the given host / port
		elastic.SetURL(fmt.Sprintf("http://%s:%d", settings.Host, settings.Port)),
	)
	if err != nil {
		return nil, err
	}

	return &ElasticV6{client: clt, config: settings}, err
}

type ElasticV6 struct {
	config *Settings
	client *elastic.Client
}

// Kill connection to remote host(s)
func (e *ElasticV6)	Close() error {
	return nil
}

// Insert collection using the given id
func (e *ElasticV6) InsertCollection(id string, in *wyc.Collection) error {
	doc := copyCollection(in)      // make copy so we don't modify the original
	doc.Id = b64encode(doc.Id)
	doc.Parent = b64encode(doc.Parent)
	doc.Name = b64encode(doc.Name) // mutate string so we aren't thrown by special chars
	doc.Uri = b64encode(doc.Uri)
	for k, v := range in.Facets {
		doc.Facets[b64encode(k)] = b64encode(v)
	}
	return e.insert(tableCollection, id, doc)
}

// Insert collection using the given id
func (e *ElasticV6) InsertItem(id string, in *wyc.Item) error {
	doc := copyItem(in) // make copy so we don't modify the original
	doc.Id = b64encode(doc.Id)
	doc.Parent = b64encode(doc.Parent)
	doc.Uri = b64encode(doc.Uri)
	// mutate user supplied strings so we aren't thrown by special chars
	doc.ItemType = b64encode(doc.ItemType)
	doc.Variant = b64encode(doc.Variant)
	for k, v := range in.Facets {
		doc.Facets[b64encode(k)] = b64encode(v)
	}
	return e.insert(tableItem, id, doc)
}

// Insert version using the given id
func (e *ElasticV6) InsertVersion(id string, in *wyc.Version) error {
	doc := copyVersion(in) // make copy so we don't modify the original
	doc.Id = b64encode(doc.Id)
	doc.Parent = b64encode(doc.Parent)
	doc.Uri = b64encode(doc.Uri)
	// mutate user supplied strings so we aren't thrown by special chars
	for k, v := range in.Facets {
		doc.Facets[b64encode(k)] = b64encode(v)
	}
	return e.insert(tableVersion, id, doc)
}

// Insert resource using the given id
func (e *ElasticV6) InsertResource(id string, in *wyc.Resource) error {
	doc := copyResource(in) // make copy so we don't modify the original
	doc.Name = b64encode(doc.Name)
	doc.Id = b64encode(doc.Id)
	doc.Parent = b64encode(doc.Parent)
	doc.ResourceType = b64encode(doc.ResourceType)
	doc.Location = b64encode(doc.Location)
	doc.Uri = b64encode(doc.Uri)
	for k, v := range in.Facets {
		doc.Facets[b64encode(k)] = b64encode(v)
	}
	return e.insert(tableResource, id, doc)
}

// Insert link using the given id
func (e *ElasticV6) InsertLink(id string, in *wyc.Link) error {
	doc := copyLink(in)
	doc.Id = b64encode(doc.Id)
	doc.Src = b64encode(doc.Src)
	doc.Dst = b64encode(doc.Dst)
	doc.Name = b64encode(doc.Name)
	doc.Uri = b64encode(doc.Uri)
	for k, v := range in.Facets {
		doc.Facets[b64encode(k)] = b64encode(v)
	}
	return e.insert(tableLink, id, doc)
}

// Update collection with given id
func (e *ElasticV6) UpdateCollection(id string, doc *wyc.Collection) error {
	// Explicit insert to ID deletes & replaces doc
	return e.InsertCollection(id, doc)
}

// Update item with given id
func (e *ElasticV6) UpdateItem(id string, doc *wyc.Item) error {
	// Explicit insert to ID deletes & replaces doc
	return e.InsertItem(id, doc)
}

// Update version with given id
func (e *ElasticV6) UpdateVersion(id string, doc *wyc.Version) error {
	// Explicit insert to ID deletes & replaces doc
	return e.InsertVersion(id, doc)
}

// Update resource with given id
func (e *ElasticV6) UpdateResource(id string, doc *wyc.Resource) error {
	// Explicit insert to ID deletes & replaces doc
	return e.InsertResource(id, doc)
}

// Update link with given id
func (e *ElasticV6) UpdateLink(id string, doc *wyc.Link) error {
	// Explicit insert to ID deletes & replaces doc
	return e.InsertLink(id, doc)
}

// Delete collections by ID(s)
func (e *ElasticV6) DeleteCollection(ids ...string) error {
	return e.delete(tableCollection, ids...)
}

// Delete items by ID(s)
func (e *ElasticV6) DeleteItem(ids ...string) error {
	return e.delete(tableItem, ids...)
}

// Delete versions by ID(s)
func (e *ElasticV6) DeleteVersion(ids ...string) error {
	return e.delete(tableVersion, ids...)
}

// Delete resources by ID(s)
func (e *ElasticV6) DeleteResource(ids ...string) error {
	return e.delete(tableResource, ids...)
}

// Delete links by ID(s)
func (e *ElasticV6) DeleteLink(ids ...string) error {
	return e.delete(tableLink, ids...)
}

// generic delete all these things
func (e *ElasticV6) delete(index string, in ...string) error {
	if len(in) < 1 {
		return nil
	}

	matchIds := make([]elastic.Query, len(in))
	for i, id := range in {
		b := elastic.NewBoolQuery()
		b.Must(elastic.NewTermQuery("Id", id)) // Id must match exactly
		matchIds[i] = b
	}

	oquery := elastic.NewBoolQuery()
	oquery.Should(matchIds...) // match any one of these Ids

	_, err := e.client.DeleteByQuery(index).Query(oquery).Do(context.Background())
	return err
}

// generic `insert this thing`
func (e *ElasticV6) insert(index, id string, in wyc.Marshalable) error {
	prep := e.client.Index().Index(index).Type(index).Id(id).BodyJson(in)

	if e.config.ReindexOnWrite {
		// If "ReindexOnWrite" is set, we'll send Elastic ?refresh=true and .. wait ..
		// https://www.elastic.co/guide/en/elasticsearch/reference/6.2/docs-refresh.html
		prep.Refresh("true")
	}

	_, err := prep.Do(context.Background())
	return err
}

// build term queries for given facets
func facetQuery(in map[string]string) (out []*elastic.TermQuery) {
	for k, v := range in {
		out = append(out, elastic.NewTermQuery(fmt.Sprintf("Facets.%s", b64encode(k)),  b64encode(v)))
	}
	return
}

// build term query for some string -> string
func termQuery(name, value string, encode bool) *elastic.TermQuery {
	if encode {
		value = b64encode(value)
	}
	return elastic.NewTermQuery(name, value)
}

// build term queries for the given query description, assuming we're looking for a collection
func collectionQuery(in *wyc.QueryDesc) (out []*elastic.TermQuery) {
	if in.Id != "" {
		out = append(out, termQuery("Id", in.Id, true))
	}
	if in.Uri != "" {
		out = append(out, termQuery("Uri", in.Uri, true))
	}
	if in.Parent != "" {
		out = append(out, termQuery("Parent", in.Parent, true))
	}
	if in.Facets != nil {
		out = append(out, facetQuery(in.Facets)...)
	}
	if in.Name != "" {
		out = append(out, termQuery("Name", in.Name, true))
	}
	return
}

// build term queries for the given query description, assuming we're looking for a item
func itemQuery(in *wyc.QueryDesc) (out []*elastic.TermQuery) {
	if in.Id != "" {
		out = append(out, termQuery("Id", in.Id, true))
	}
	if in.Uri != "" {
		out = append(out, termQuery("Uri", in.Uri, true))
	}
	if in.Parent != "" {
		out = append(out, termQuery("Parent", in.Parent, true))
	}
	if in.Facets != nil {
		out = append(out, facetQuery(in.Facets)...)
	}
	if in.ItemType != "" {
		out = append(out, termQuery("ItemType", in.ItemType, true))
	}
	if in.Variant != "" {
		out = append(out, termQuery("Variant", in.Variant, true))
	}
	return
}

// build term queries for the given query description, assuming we're looking for a version
func versionQuery(in *wyc.QueryDesc) (out []*elastic.TermQuery) {
	if in.Id != "" {
		out = append(out, termQuery("Id", in.Id, true))
	}
	if in.Uri != "" {
		out = append(out, termQuery("Uri", in.Uri, true))
	}
	if in.Parent != "" {
		out = append(out, termQuery("Parent", in.Parent, true))
	}
	if in.Facets != nil {
		out = append(out, facetQuery(in.Facets)...)
	}
	if in.VersionNumber > 0 {
		out = append(out, termQuery("Number", fmt.Sprintf("%d", in.VersionNumber), false))
	}
	return
}

// build term queries for the given query description, assuming we're looking for a resource
func resourceQuery(in *wyc.QueryDesc) (out []*elastic.TermQuery) {
	if in.Id != "" {
		out = append(out, termQuery("Id", in.Id, true))
	}
	if in.Uri != "" {
		out = append(out, termQuery("Uri", in.Uri, true))
	}
	if in.Parent != "" {
		out = append(out, termQuery("Parent", in.Parent, true))
	}
	if in.Facets != nil {
		out = append(out, facetQuery(in.Facets)...)
	}
	if in.Name != "" {
		out = append(out, termQuery("Name", in.Name, true))
	}
	if in.Location != "" {
		out = append(out, termQuery("Location", in.Location, true))
	}
	if in.ResourceType != "" {
		out = append(out, termQuery("ResourceType", in.ResourceType, true))
	}
	return
}

// build term queries for the given query description, assuming we're looking for a link
func linkQuery(in *wyc.QueryDesc) (out []*elastic.TermQuery) {
	if in.Id != "" {
		out = append(out, termQuery("Id", in.Id, true))
	}
	if in.Uri != "" {
		out = append(out, termQuery("Uri", in.Uri, true))
	}
	if in.Facets != nil {
		out = append(out, facetQuery(in.Facets)...)
	}
	if in.Name != "" {
		out = append(out, termQuery("Name", in.Name, true))
	}
	if in.LinkSrc != "" {
		out = append(out, termQuery("Src", in.LinkSrc, true))
	}
	if in.LinkDst != "" {
		out = append(out, termQuery("Dst", in.LinkDst, true))
	}
	return
}

// Given some term queries, join them with 'Must' and return the query that contains them in a 'Should'.
//
func joinAnd(in ...*elastic.TermQuery) *elastic.BoolQuery {
	must := elastic.NewBoolQuery()
	for _, tq := range in {
		must.Must(tq)
	}
	return must
}

type converter func (*wyc.QueryDesc) ([]*elastic.TermQuery)

// generic query func
//  string: index to check
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  []*wyc.QueryDesc: queries to build from
//  converter: function to use to convert QueryDesc -> elastic.Query
func (e *ElasticV6) query(index string, limit, offset int, queries []*wyc.QueryDesc, fn converter) ([]string, error) {
	var q elastic.Query

	if len(queries) == 0 {
		// Special case if there are no given QueryDesc items
		q = elastic.NewMatchAllQuery()
	} else {
		// Turn QueryDesc(s) into elastic Query(s)
		tmp := elastic.NewBoolQuery()
		for _, query := range queries {
			terms := fn(query)
			if len(terms) == 0 {
				continue
			}
			tmp.Should(joinAnd(terms...))
		}
		q = tmp
	}

	// prep query
	base := e.client.Search().From(offset).Index(index).Type(index).NoStoredFields()
	base.Size(matchAllSearchLimit) // default limit

	// be sure to set some kind of limit
	if limit > 0 && limit <= matchAllSearchLimit {
		base.Size(limit)
	}

	hits, err := base.Query(q).Do(context.Background())
	if err != nil {
		exists, eerr := e.client.IndexExists(index).Do(context.Background())
		if eerr == nil && !exists {
			return make([]string, 0), nil
		}
		return nil, err
	}

	results := make([]string, hits.Hits.TotalHits)
	if len(results) == 0 {
		return results, nil
	}
	for i, hit := range hits.Hits.Hits {
		results[i] = hit.Id
	}

	return results, nil
}

// Query for collections
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  ...QueryDes: description(s) of what to search for
// Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
func (e *ElasticV6)	QueryCollection(limit, offset int, in ...*wyc.QueryDesc) ([]string, error) {
	return e.query(tableCollection, limit, offset, in, collectionQuery)
}

// Query for items
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  ...QueryDes: description(s) of what to search for
// Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
func (e *ElasticV6)	QueryItem(limit , offset int, in ...*wyc.QueryDesc) ([]string, error) {
	return e.query(tableItem, limit, offset, in, itemQuery)
}

// Query for versions
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  ...QueryDes: description(s) of what to search for
// Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
func (e *ElasticV6)	QueryVersion(limit , offset int, in ...*wyc.QueryDesc) ([]string, error) {
	return e.query(tableVersion, limit, offset, in, versionQuery)
}

// Query for resources
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  ...QueryDes: description(s) of what to search for
// Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
func (e *ElasticV6)	QueryResource(limit , offset int, in ...*wyc.QueryDesc) ([]string, error) {
	return e.query(tableResource, limit, offset, in, resourceQuery)
}

// Query for links
//  int: 'limit' results to at most int (where 0 indicates there is no limit)
//  int: 'from' what number to start returning results from
//  ...QueryDes: description(s) of what to search for
// Ids will be returned for any doc matching all of the fields of any of the given QueryDesc
func (e *ElasticV6)	QueryLink(limit , offset int, in ...*wyc.QueryDesc) ([]string, error) {
	return e.query(tableLink, limit, offset, in, linkQuery)
}
