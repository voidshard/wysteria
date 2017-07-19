package searchends

import (
	"fmt"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	wyc "github.com/voidshard/wysteria/common"
	"os"
	"strings"
)

// Wrapper struct around bleve search indexes
type bleveSearchbase struct {
	collections bleve.Index
	items       bleve.Index
	versions    bleve.Index
	resources   bleve.Index
	links       bleve.Index
}

// Return a new bleve index given a path on disk, either by creating if it doesn't exist or opening it if it does.
func newBleveIndex(name string, documentMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	_, err := os.Stat(name)
	if err == nil {
		// Open existing index
		return bleve.Open(name)
	}
	// Create new index
	return bleve.New(name, documentMapping)
}

// 'connect' to bleve by opening/creating all of our indexes
func bleveConnect(settings *Settings) (Searchbase, error) {
	sb := &bleveSearchbase{}
	imapping := bleve.NewIndexMapping()

	idx, err := newBleveIndex(settings.Database+tableCollection, imapping)
	if err != nil {
		return nil, err
	}
	sb.collections = idx

	idx, err = newBleveIndex(settings.Database+tableItem, imapping)
	if err != nil {
		return nil, err
	}
	sb.items = idx

	idx, err = newBleveIndex(settings.Database+tableVersion, imapping)
	if err != nil {
		return nil, err
	}
	sb.versions = idx

	idx, err = newBleveIndex(settings.Database+tableResource, imapping)
	if err != nil {
		return nil, err
	}
	sb.resources = idx

	idx, err = newBleveIndex(settings.Database+tableLink, imapping)
	if err != nil {
		return nil, err
	}
	sb.links = idx

	return sb, nil
}

// close bleve connections (or more accurately, close open filehandlers I guess)
func (b *bleveSearchbase) Close() error {
	b.collections.Close()
	b.items.Close()
	b.versions.Close()
	b.resources.Close()
	b.links.Close()
	return nil
}

// Insert collection using the given id
func (b *bleveSearchbase) InsertCollection(id string, doc *wyc.Collection) error {
	in := copyCollection(doc)    // create copy so we don't modify the original
	in.Name = b64encode(in.Name) // encode name so we aren't tripped up by spaces / bleve control chars
	return b.collections.Index(id, in)
}

// Insert collection using the given id
func (b *bleveSearchbase) InsertItem(id string, doc *wyc.Item) error {
	in := copyItem(doc) // create copy so we don't modify the original

	// mutate values of our copy so we don't have to worry about weird chars
	in.ItemType = b64encode(in.ItemType)
	in.Variant = b64encode(in.Variant)
	for k, v := range doc.Facets {
		in.Facets[b64encode(k)] = b64encode(v)
	}

	return b.items.Index(id, in)
}

// Insert version using the given id
func (b *bleveSearchbase) InsertVersion(id string, doc *wyc.Version) error {
	in := copyVersion(doc) // create copy so we don't modify the original

	// mutate values of our copy so we don't have to worry about weird chars
	for k, v := range doc.Facets {
		in.Facets[b64encode(k)] = b64encode(v)
	}
	return b.versions.Index(id, in)
}

// Insert resource using the given id
func (b *bleveSearchbase) InsertResource(id string, doc *wyc.Resource) error {
	in := copyResource(doc)
	in.Name = b64encode(in.Name)
	in.ResourceType = b64encode(in.ResourceType)
	in.Location = b64encode(in.Location)
	return b.resources.Index(id, in)
}

// Insert link using the given id
func (b *bleveSearchbase) InsertLink(id string, doc *wyc.Link) error {
	in := copyLink(doc)
	in.Name = b64encode(in.Name)
	return b.links.Index(id, in)
}

// Update item with given id
func (b *bleveSearchbase) UpdateItem(id string, in *wyc.Item) error {
	// For bleve, updating and inserting are the same thing
	return b.InsertItem(id, in)
}

// Update item with given id
func (b *bleveSearchbase) UpdateVersion(id string, in *wyc.Version) error {
	// For bleve, updating and inserting are the same thing
	return b.InsertVersion(id, in)
}

// Iterate over given IDs and delete from given index
func genericDelete(index bleve.Index, ids ...string) error {
	for _, id := range ids {
		err := index.Delete(id)
		if err != nil {
			return err
		}
	}
	return nil
}

// Delete collections by ID(s)
func (b *bleveSearchbase) DeleteCollection(ids ...string) error {
	return genericDelete(b.collections, ids...)
}

// Delete items by ID(s)
func (b *bleveSearchbase) DeleteItem(ids ...string) error {
	return genericDelete(b.items, ids...)
}

// Delete versions by ID(s)
func (b *bleveSearchbase) DeleteVersion(ids ...string) error {
	return genericDelete(b.versions, ids...)
}

// Delete resources by ID(s)
func (b *bleveSearchbase) DeleteResource(ids ...string) error {
	return genericDelete(b.resources, ids...)
}

// Delete links by ID(s)
func (b *bleveSearchbase) DeleteLink(ids ...string) error {
	return genericDelete(b.links, ids...)
}

// Transform a QueryDesc into a bleve compatible search string for a collection
func toCollectionQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", b64encode(desc.Name)))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	return strings.Join(sq, " ")
}

// Transform a QueryDesc into a bleve compatible search string for a item
func toItemQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.ItemType != "" {
		sq = append(sq, fmt.Sprintf("+ItemType:%s", b64encode(desc.ItemType)))
	}
	if desc.Variant != "" {
		sq = append(sq, fmt.Sprintf("+Variant:%s", b64encode(desc.Variant)))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	for k, v := range desc.Facets {
		sq = append(sq, fmt.Sprintf("+Facets.%s:%s", b64encode(k), b64encode(v)))
	}
	return strings.Join(sq, " ")
}

// Transform a QueryDesc into a bleve compatible search string for a version
func toVersionQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	for k, v := range desc.Facets {
		sq = append(sq, fmt.Sprintf("+Facets.%s:%s", b64encode(k), b64encode(v)))
	}
	if desc.VersionNumber > 0 {
		sq = append(sq, fmt.Sprintf("+Number:%d", desc.VersionNumber))
	}
	return strings.Join(sq, " ")
}

// Transform a QueryDesc into a bleve compatible search string for a resource
func toResourceQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	if desc.ResourceType != "" {
		sq = append(sq, fmt.Sprintf("+ResourceType:%s", b64encode(desc.ResourceType)))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", b64encode(desc.Name)))
	}
	if desc.Location != "" {
		sq = append(sq, fmt.Sprintf("+Location:%s", b64encode(desc.Location)))
	}
	return strings.Join(sq, " ")
}

// Transform a QueryDesc into a bleve compatible search string for a link
func toLinkQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", b64encode(desc.Name)))
	}
	if desc.LinkSrc != "" {
		sq = append(sq, fmt.Sprintf("+Src:%s", desc.LinkSrc))
	}
	if desc.LinkDst != "" {
		sq = append(sq, fmt.Sprintf("+Dst:%s", desc.LinkDst))
	}
	return strings.Join(sq, " ")
}

// A generic query function that handles grabbing IDs from a bleve index given
//  - limit: max number of entries to return
//  - from: return only entries after this number
//  - index: a bleve index to search
//  - convert: a function to take a QueryDesc and turn it into a bleve query string
//  - queries: the search QueryDesc objects
func genericQuery(limit, from int, index bleve.Index, convert func(desc *wyc.QueryDesc) string, queries ...*wyc.QueryDesc) ([]string, error) {
	// ToDo: There is probably a smarter way to do this as a single query with limit / page

	if len(queries) < 1 {
		return nil, nil
	}

	var result *bleve.SearchResult
	for _, query := range queries {
		search_query := bleve.NewQueryStringQuery(convert(query))

		search := bleve.NewSearchRequestOptions(search_query, limit, from, false)
		res, err := index.Search(search)

		if err != nil {
			return nil, err
		}
		if result == nil {
			result = res
		} else {
			if res.Hits.Len() > 0 {
				result.Merge(res)
			}
		}
	}

	ids := []string{}
	for _, doc := range result.Hits {
		ids = append(ids, doc.ID)
	}

	return ids, nil
}

// Special case 'match all' query
//
func (b *bleveSearchbase) emptyQuery(limit, from int, index bleve.Index) ([]string, error) {
	if limit > matchAllSearchLimit {
		limit = matchAllSearchLimit
	}

	search_query := bleve.NewQueryStringQuery("*")
	search := bleve.NewSearchRequestOptions(search_query, limit, from, false)
	result, err := index.Search(search)

	if err != nil || result.Hits.Len() == 0 {
		return nil, err
	}

	ids := []string{}
	for _, doc := range result.Hits {
		ids = append(ids, doc.ID)
	}
	return ids, nil
}

// Search for collections matching the given query descriptions
func (b *bleveSearchbase) QueryCollection(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	if len(query) == 0 {
		return b.emptyQuery(limit, from, b.collections)
	}
	return genericQuery(limit, from, b.collections, toCollectionQueryString, query...)
}

// Search for items matching the given query descriptions
func (b *bleveSearchbase) QueryItem(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	if len(query) == 0 {
		return b.emptyQuery(limit, from, b.items)
	}
	return genericQuery(limit, from, b.items, toItemQueryString, query...)
}

// Search for versions matching the given query descriptions
func (b *bleveSearchbase) QueryVersion(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	if len(query) == 0 {
		return b.emptyQuery(limit, from, b.versions)
	}
	return genericQuery(limit, from, b.versions, toVersionQueryString, query...)
}

// Search for resources matching the given query descriptions
func (b *bleveSearchbase) QueryResource(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	if len(query) == 0 {
		return b.emptyQuery(limit, from, b.resources)
	}
	return genericQuery(limit, from, b.resources, toResourceQueryString, query...)
}

// Search for links matching the given query descriptions
func (b *bleveSearchbase) QueryLink(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	if len(query) == 0 {
		return b.emptyQuery(limit, from, b.links)
	}
	return genericQuery(limit, from, b.links, toLinkQueryString, query...)
}
