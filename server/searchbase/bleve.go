package searchends

import (
	wyc "github.com/voidshard/wysteria/common"
	"github.com/blevesearch/bleve"
	"github.com/blevesearch/bleve/mapping"
	"fmt"
	"strings"
	"os"
)

type bleveSearchbase struct {
	collections bleve.Index
	items bleve.Index
	versions bleve.Index
	resources bleve.Index
	links bleve.Index
}

func create_bleve_index(name string, documentMapping *mapping.IndexMappingImpl) (bleve.Index, error) {
	_, err := os.Stat(name)
	if err == nil {
		// Open existing index
		return bleve.Open(name)
	}
	// Create new index
	return bleve.New(name, documentMapping)
}

func bleve_connect(settings *SearchbaseSettings) (Searchbase, error) {
	sb := &bleveSearchbase{}
	imapping := bleve.NewIndexMapping()

	idx, err := create_bleve_index(settings.Database + table_collection, imapping)
	if err != nil {
		return nil, err
	}
	sb.collections = idx

	idx, err = create_bleve_index(settings.Database + table_item, imapping)
	if err != nil {
		return nil, err
	}
	sb.items = idx

	idx, err = create_bleve_index(settings.Database + table_version, imapping)
	if err != nil {
		return nil, err
	}
	sb.versions = idx

	idx, err = create_bleve_index(settings.Database + table_fileresource, imapping)
	if err != nil {
		return nil, err
	}
	sb.resources = idx

	idx, err = create_bleve_index(settings.Database + table_link, imapping)
	if err != nil {
		return nil, err
	}
	sb.links = idx

	return sb, nil
}

func (b *bleveSearchbase) Close() error {
	return nil
}

func (b *bleveSearchbase) InsertCollection(id string, in *wyc.Collection) error {
	return b.collections.Index(id, in)
}

func (b *bleveSearchbase) InsertItem(id string, in *wyc.Item) error {
	return b.items.Index(id, in)
}

func (b *bleveSearchbase) InsertVersion(id string, in *wyc.Version) error {
	return b.versions.Index(id, in)
}

func (b *bleveSearchbase) InsertResource(id string, in *wyc.Resource) error {
	in.Location = b64encode(in.Location)
	return b.resources.Index(id, in)
}

func (b *bleveSearchbase) InsertLink(id string, in *wyc.Link) error {
	return b.links.Index(id, in)
}

func (b *bleveSearchbase) UpdateItem(id string, in *wyc.Item) error {
	return b.items.Index(id, in)
}

func (b *bleveSearchbase) UpdateVersion(id string, in *wyc.Version) error {
	return b.versions.Index(id, in)
}

func generic_delete (index bleve.Index, ids ...string) error {
	for _, id := range ids {
		err := index.Delete(id)
		if err != nil {
			return err
		}
	}
	return nil
}

func (b *bleveSearchbase) DeleteCollection(ids ...string) error {
	return generic_delete(b.collections, ids...)
}

func (b *bleveSearchbase) DeleteItem(ids ...string) error {
	return generic_delete(b.items, ids...)
}

func (b *bleveSearchbase) DeleteVersion(ids ...string) error {
	return generic_delete(b.versions, ids...)
}

func (b *bleveSearchbase) DeleteResource(ids ...string) error {
	return generic_delete(b.resources, ids...)
}

func (b *bleveSearchbase) DeleteLink(ids ...string) error {
	return generic_delete(b.links, ids...)
}

func toCollectionQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", desc.Name))
	}
	return strings.Join(sq, " ")
}

func toItemQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.ItemType != "" {
		sq = append(sq, fmt.Sprintf("+ItemType:%s", desc.ItemType))
	}
	if desc.Variant != "" {
		sq = append(sq, fmt.Sprintf("+Variant:%s", desc.Variant))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	for k, v := range desc.Facets {
		sq = append(sq, fmt.Sprintf("%s", k), v)
	}
	return strings.Join(sq, " ")
}

func toVersionQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	for k, v := range desc.Facets {
		sq = append(sq, fmt.Sprintf("%s", k), v)
	}
	if desc.VersionNumber > 0 {
		sq = append(sq, fmt.Sprintf("+Number:%d", desc.VersionNumber))
	}
	return strings.Join(sq, " ")
}

func toResourceQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Parent != "" {
		sq = append(sq, fmt.Sprintf("+Parent:%s", desc.Parent))
	}
	if desc.ResourceType != "" {
		sq = append(sq, fmt.Sprintf("+ResourceType:%s", desc.ResourceType))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", desc.Name))
	}
	if desc.Location != "" {
		hsh := b64encode(desc.Location)
		sq = append(sq, fmt.Sprintf("+Location:%s", hsh))
	}
	return strings.Join(sq, " ")
}

func toLinkQueryString(desc *wyc.QueryDesc) string {
	sq := []string{}
	if desc.Id != "" {
		sq = append(sq, fmt.Sprintf("+Id:%s", desc.Id))
	}
	if desc.Name != "" {
		sq = append(sq, fmt.Sprintf("+Name:%s", desc.Name))
	}
	if desc.LinkSrc != "" {
		sq = append(sq, fmt.Sprintf("+Src:%s", desc.LinkSrc))
	}
	if desc.LinkDst != "" {
		sq = append(sq, fmt.Sprintf("+Dest:%s", desc.LinkDst))
	}
	return strings.Join(sq, " ")
}

func genericQuery(limit, from int, index bleve.Index, convert func(desc *wyc.QueryDesc) string, queries ...*wyc.QueryDesc) ([]string, error) {
	// ToDo: There is probably a smarter way to do this as a single query with limit / page

	var result *bleve.SearchResult
	for _, query := range queries {
		search_query := bleve.NewQueryStringQuery(convert(query))
		search := bleve.NewSearchRequest(search_query)
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

	if limit < 1 {
		// there is no limit, return them all
		return ids, nil
	}

	if limit >= len(ids) {
		// we found less results than our limit, return them all
		return ids, nil
	}

	if limit + from >= len(ids) {
		// we've been asked for the last segment of the values
		return ids[from:], nil
	}

	// We've been asked for a page of results somewhere in the middle
	return ids[from: limit + from], nil
}

func (b *bleveSearchbase) QueryCollection(limit, from int, query ...*wyc.QueryDesc) ([]string, error) {
	return genericQuery(limit, from, b.collections, toCollectionQueryString, query...)
}

func (b *bleveSearchbase) QueryItem(limit , from int, query ...*wyc.QueryDesc) ([]string, error) {
	return genericQuery(limit, from, b.items, toItemQueryString, query...)
}

func (b *bleveSearchbase) QueryVersion(limit , from int, query ...*wyc.QueryDesc) ([]string, error) {
	return genericQuery(limit, from, b.versions, toVersionQueryString, query...)
}

func (b *bleveSearchbase) QueryResource(limit , from int, query ...*wyc.QueryDesc) ([]string, error) {
	return genericQuery(limit, from, b.resources, toResourceQueryString, query...)
}

func (b *bleveSearchbase) QueryLink(limit , from int, query ...*wyc.QueryDesc) ([]string, error) {
	return genericQuery(limit, from, b.links, toLinkQueryString, query...)
}
