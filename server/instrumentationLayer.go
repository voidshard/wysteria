/*
As the name 'shim' implies, ths layer sits between the bottom of the middleware and the server itself,
it supplies no business logic at all (nor should it) - no attempt here is made to address errors or
change data, it only logs traffic & events as they pass by.
*/

package main

import (
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
	wym "github.com/voidshard/wysteria/common/middleware"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	"time"
)

type Shim struct {
	server *WysteriaServer
}

// The server calls us to listen, we'll call the middleware server in turn.
// Essentially, all we have to do is pass back and forth, and be sure not to change any state.
//
func (s *Shim) ListenAndServe(config *wym.Settings, server *WysteriaServer) error {
	s.server = server
	return server.middleware_server.ListenAndServe(config, s)
}

// Time is up, kill everything and shutdown the server, kill all connections
//
func (s *Shim) Shutdown() error {
	s.server.Shutdown()
	return nil
}

// Call on our monitor to do the actual logging business
//
func (s *Shim) log(err error, t int64, opts ...wsi.Opt) {
	go func() {
		opts = append(opts, wsi.Time(time.Now().UnixNano()-t))
		if err != nil {
			s.server.monitor.Err(err, opts...)
			return
		}
		s.server.monitor.Log("", opts...)
	}()
}

// All following funcs simply record the current time, call the appropriate server call
// and pass the results back to the calling middleware.
// Except of course, we log the call in the middle.
//
func (s *Shim) CreateCollection(in *wyc.Collection) (string, error) {
	ts := time.Now().UnixNano()
	result, err := s.server.CreateCollection(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetCollection(), wsi.Note(in.Name))
	return result, err
}

func (s *Shim) CreateItem(in *wyc.Item) (string, error) {
	ts := time.Now().UnixNano()
	result, err := s.server.CreateItem(in)
	note := fmt.Sprintf("%s %s %s", in.Parent, in.ItemType, in.Variant)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetItem(), wsi.Note(note))
	return result, err
}

func (s *Shim) CreateVersion(in *wyc.Version) (string, int32, error) {
	ts := time.Now().UnixNano()
	resultId, resultVer, err := s.server.CreateVersion(in)
	note := fmt.Sprintf("%s %s %d", in.Parent, resultId, resultVer)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(note))
	return resultId, resultVer, err
}

func (s *Shim) CreateResource(in *wyc.Resource) (string, error) {
	ts := time.Now().UnixNano()
	result, err := s.server.CreateResource(in)
	note := fmt.Sprintf("%s %s %s %s", in.Parent, in.Name, in.ResourceType, in.Location)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetResource(), wsi.Note(note))
	return result, err
}

func (s *Shim) CreateLink(in *wyc.Link) (string, error) {
	ts := time.Now().UnixNano()
	result, err := s.server.CreateLink(in)
	note := fmt.Sprintf("%s %s", in.Src, in.Dst)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetLink(), wsi.Note(note))
	return result, err
}

func (s *Shim) DeleteCollection(in string) error {
	ts := time.Now().UnixNano()
	err := s.server.DeleteCollection(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetCollection(), wsi.Note(in))
	return err
}

func (s *Shim) DeleteItem(in string) error {
	ts := time.Now().UnixNano()
	err := s.server.DeleteItem(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetItem(), wsi.Note(in))
	return err
}

func (s *Shim) DeleteVersion(in string) error {
	ts := time.Now().UnixNano()
	err := s.server.DeleteVersion(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetVersion(), wsi.Note(in))
	return err
}

func (s *Shim) DeleteResource(in string) error {
	ts := time.Now().UnixNano()
	err := s.server.DeleteResource(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetResource(), wsi.Note(in))
	return err
}

func (s *Shim) FindCollections(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	ts := time.Now().UnixNano()
	results, err := s.server.FindCollections(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetCollection(), wsi.Note(fmt.Sprintf("%d %d %d", l, o, len(q))))
	return results, err
}

func (s *Shim) FindItems(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Item, error) {
	ts := time.Now().UnixNano()
	results, err := s.server.FindItems(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetItem(), wsi.Note(fmt.Sprintf("%d %d %d", l, o, len(q))))
	return results, err
}

func (s *Shim) FindVersions(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Version, error) {
	ts := time.Now().UnixNano()
	results, err := s.server.FindVersions(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetVersion(), wsi.Note(fmt.Sprintf("%d %d %d", l, o, len(q))))
	return results, err
}

func (s *Shim) FindResources(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	ts := time.Now().UnixNano()
	results, err := s.server.FindResources(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetResource(), wsi.Note(fmt.Sprintf("%d %d %d", l, o, len(q))))
	return results, err
}

func (s *Shim) FindLinks(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Link, error) {
	ts := time.Now().UnixNano()
	results, err := s.server.FindLinks(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetLink(), wsi.Note(fmt.Sprintf("%d %d %d", l, o, len(q))))
	return results, err
}

func (s *Shim) PublishedVersion(in string) (*wyc.Version, error) {
	ts := time.Now().UnixNano()
	result, err := s.server.PublishedVersion(in)
	s.log(err, ts, wsi.IsPublished(), wsi.TargetVersion(), wsi.Note(in))
	return result, err
}

func (s *Shim) SetPublishedVersion(in string) error {
	ts := time.Now().UnixNano()
	err := s.server.SetPublishedVersion(in)
	s.log(err, ts, wsi.IsPublish(), wsi.TargetVersion(), wsi.Note(in))
	return err
}

func (s *Shim) UpdateVersionFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	err := s.server.UpdateVersionFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetVersion(), wsi.Note(in))
	return err
}

func (s *Shim) UpdateItemFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	err := s.server.UpdateItemFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(in))
	return err
}

func (s *Shim) UpdateCollectionFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	err := s.server.UpdateCollection(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(in))
	return err
}

func (s *Shim) UpdateResourceFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	err := s.server.UpdateResource(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(in))
	return err
}

func (s *Shim) UpdateLinkFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	err := s.server.UpdateResource(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(in))
	return err
}
