/*
Database instrumentation
 - This is an implementation of Database that accepts a logging monitor and another implementor
   of Database. All this does is pass any calls to the database it wraps & passes
   back the result(s). In the process, every action is logged to the given monitor.
*/

package main

import (
	wyc "github.com/voidshard/wysteria/common"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	wyd "github.com/voidshard/wysteria/server/searchbase"
	"time"
)

func newSearchbaseMonitor(db wyd.Searchbase, monitor *wsi.Monitor) wyd.Searchbase {
	return &SearchbaseMonitor{searchbase: db, monitor: monitor}
}

type SearchbaseMonitor struct {
	searchbase wyd.Searchbase
	monitor    *wsi.Monitor
}

// Call on our monitor to do the actual logging business
//
func (s *SearchbaseMonitor) log(err error, t int64, opts ...wsi.Opt) {
	area := "searchbase"
	go func() {
		opts = append(opts, wsi.Time(time.Now().UnixNano()-t))
		if err != nil {
			s.monitor.Err(err, area, opts...)
			return
		}
		s.monitor.Log(area, opts...)
	}()
}

// Kill connection to remote host(s)
func (s *SearchbaseMonitor) Close() error {
	return s.Close()
}

// Insert collection into the sb with the given Id
func (s *SearchbaseMonitor) InsertCollection(id string, in *wyc.Collection) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.InsertCollection(id, in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetCollection(), wsi.Note(in.Parent, in.Id, in.Name))
	return err
}

// Insert item into the sb with the given Id
func (s *SearchbaseMonitor) InsertItem(id string, in *wyc.Item) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.InsertItem(id, in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetItem(), wsi.Note(in.Parent, in.Id, in.ItemType, in.Variant))
	return err
}

// Insert version into the sb with the given Id
func (s *SearchbaseMonitor) InsertVersion(id string, in *wyc.Version) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.InsertVersion(id, in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(in.Parent, in.Id, in.Number))
	return err
}

// Insert resource into the sb with the given Id
func (s *SearchbaseMonitor) InsertResource(id string, in *wyc.Resource) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.InsertResource(id, in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(in.Parent, in.Id, in.Name, in.ResourceType, in.Location))
	return err
}

// Insert link into the sb with the given Id
func (s *SearchbaseMonitor) InsertLink(id string, in *wyc.Link) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.InsertLink(id, in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(in.Id, in.Src, in.Dst))
	return err
}

// Update the facets of the collection with the given id with the given facets
func (s *SearchbaseMonitor) UpdateCollection(id string, in *wyc.Collection) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.UpdateCollection(id, in)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetCollection(), wsi.Note(id))
	return err
}

// Update the facets of the resource with the given id with the given facets
func (s *SearchbaseMonitor) UpdateResource(id string, in *wyc.Resource) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.UpdateResource(id, in)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetResource(), wsi.Note(id))
	return err
}

// Update the facets of the link with the given id with the given facets
func (s *SearchbaseMonitor) UpdateLink(id string, in *wyc.Link) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.UpdateLink(id, in)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetLink(), wsi.Note(id))
	return err
}

// Update the facets of the item with the given id with the given facets
func (s *SearchbaseMonitor) UpdateItem(id string, in *wyc.Item) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.UpdateItem(id, in)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(id))
	return err
}

// Update the facets of the version with the given id with the given facets
func (s *SearchbaseMonitor) UpdateVersion(id string, in *wyc.Version) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.UpdateVersion(id, in)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetVersion(), wsi.Note(id))
	return err
}

// Delete collection search data by Id(s)
func (s *SearchbaseMonitor) DeleteCollection(in ...string) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.DeleteCollection(in...)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetCollection(), wsi.Note(in))
	return err
}

// Delete item search data by Id(s)
func (s *SearchbaseMonitor) DeleteItem(in ...string) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.DeleteItem(in...)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetItem(), wsi.Note(in))
	return err
}

// Delete version search data by Id(s)
func (s *SearchbaseMonitor) DeleteVersion(in ...string) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.DeleteVersion(in...)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetVersion(), wsi.Note(in))
	return err
}

// Delete resource search data by Id(s)
func (s *SearchbaseMonitor) DeleteResource(in ...string) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.DeleteResource(in...)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetResource(), wsi.Note(in))
	return err
}

// Delete link search data by Id(s)
func (s *SearchbaseMonitor) DeleteLink(in ...string) error {
	ts := time.Now().UnixNano()
	err := s.searchbase.DeleteLink(in...)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetLink(), wsi.Note(in))
	return err
}

// Query for collections
func (s *SearchbaseMonitor) QueryCollection(l, o int, q ...*wyc.QueryDesc) ([]string, error) {
	ts := time.Now().UnixNano()
	results, err := s.searchbase.QueryCollection(l, o, q...)
	s.log(err, ts, wsi.IsFind(), wsi.TargetCollection(), wsi.Note(l, o, len(results)))
	return results, err
}

// Query for items
func (s *SearchbaseMonitor) QueryItem(l, o int, q ...*wyc.QueryDesc) ([]string, error) {
	ts := time.Now().UnixNano()
	results, err := s.searchbase.QueryItem(l, o, q...)
	s.log(err, ts, wsi.IsFind(), wsi.TargetItem(), wsi.Note(l, o, len(results)))
	return results, err
}

// Query for versions
func (s *SearchbaseMonitor) QueryVersion(l, o int, q ...*wyc.QueryDesc) ([]string, error) {
	ts := time.Now().UnixNano()
	results, err := s.searchbase.QueryVersion(l, o, q...)
	s.log(err, ts, wsi.IsFind(), wsi.TargetVersion(), wsi.Note(l, o, len(results)))
	return results, err
}

// Query for resources
func (s *SearchbaseMonitor) QueryResource(l, o int, q ...*wyc.QueryDesc) ([]string, error) {
	ts := time.Now().UnixNano()
	results, err := s.searchbase.QueryResource(l, o, q...)
	s.log(err, ts, wsi.IsFind(), wsi.TargetResource(), wsi.Note(l, o, len(results)))
	return results, err
}

// Query for links
func (s *SearchbaseMonitor) QueryLink(l, o int, q ...*wyc.QueryDesc) ([]string, error) {
	ts := time.Now().UnixNano()
	results, err := s.searchbase.QueryLink(l, o, q...)
	s.log(err, ts, wsi.IsFind(), wsi.TargetLink(), wsi.Note(l, o, len(results)))
	return results, err
}
