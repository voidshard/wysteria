/*
The MiddlewareMonitor layer sits between the bottom of the middleware and the server itself,
it supplies no business logic at all (nor should it) - no attempt here is made to address errors or
change data, it only logs traffic & events as they pass by.
*/

package main

import (
	wyc "github.com/voidshard/wysteria/common"
	wym "github.com/voidshard/wysteria/common/middleware"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	"time"
)

const (
	keywordEnter = "enter"
	keywordExit = "exit"
)

type MiddlewareMonitor struct {
	// This is the actual server handler that will be called - amidst our logging
	server wym.ServerHandler

	// The middleware endpoint to kick off (we'll take input from here and pass them back & forth)
	epServer wym.EndpointServer

	// The monitor to use to log things to
	monitor *wsi.Monitor
}

func newMiddlewareMonitor(endpoint wym.EndpointServer, monitor *wsi.Monitor) *MiddlewareMonitor {
	return &MiddlewareMonitor{
		epServer: endpoint,
		monitor: monitor,
	}
}

// The server calls us to listen, we'll call the middleware server in turn.
// Essentially, all we have to do is pass back and forth, and be sure not to change any state.
//
func (s *MiddlewareMonitor) ListenAndServe(config *wym.Settings, server wym.ServerHandler) error {
	s.server = server
	return s.epServer.ListenAndServe(config, s)
}

// Time is up, kill everything and shutdown the server, kill all connections
//
func (s *MiddlewareMonitor) Shutdown() error {
	s.epServer.Shutdown()
	return nil
}

// Call on our monitor to do the actual logging business
//
func (s *MiddlewareMonitor) log(err error, t int64, opts ...wsi.Opt) {
	area := "middleware"
	go func() {
		opts = append(opts, wsi.Time(time.Now().UnixNano() - t))
		if err != nil {
			s.monitor.Err(err, area, opts...)
			return
		}
		s.monitor.Log(area, opts...)
	}()
}

// All following funcs simply record the current time, call the appropriate server call
// and pass the results back to the calling middleware.
// Except of course, we log the call in the middle.
//
func (s *MiddlewareMonitor) CreateCollection(in *wyc.Collection) (string, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsCreate(), wsi.TargetCollection(), wsi.Note(keywordEnter, in.Name))
	result, err := s.server.CreateCollection(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetCollection(), wsi.Note(keywordExit, in.Name))
	return result, err
}

func (s *MiddlewareMonitor) CreateItem(in *wyc.Item) (string, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsCreate(), wsi.TargetItem(), wsi.Note(keywordEnter, in.Parent, in.Id, in.ItemType, in.Variant))
	result, err := s.server.CreateItem(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetItem(), wsi.Note(keywordExit, in.Parent, in.Id, in.ItemType, in.Variant))
	return result, err
}

func (s *MiddlewareMonitor) CreateVersion(in *wyc.Version) (string, int32, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(keywordEnter, in.Parent))
	resultId, resultVer, err := s.server.CreateVersion(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetVersion(), wsi.Note(keywordExit, in.Parent, resultId, resultVer))
	return resultId, resultVer, err
}

func (s *MiddlewareMonitor) CreateResource(in *wyc.Resource) (string, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsCreate(), wsi.TargetResource(), wsi.Note(keywordEnter, in.Parent, in.Name, in.ResourceType, in.Location))
	result, err := s.server.CreateResource(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetResource(), wsi.Note(keywordExit, in.Parent, in.Name, in.ResourceType, in.Location))
	return result, err
}

func (s *MiddlewareMonitor) CreateLink(in *wyc.Link) (string, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsCreate(), wsi.TargetLink(), wsi.Note(keywordEnter, in.Src, in.Dst))
	result, err := s.server.CreateLink(in)
	s.log(err, ts, wsi.IsCreate(), wsi.TargetLink(), wsi.Note(keywordExit, in.Src, in.Dst))
	return result, err
}

func (s *MiddlewareMonitor) DeleteCollection(in string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsDelete(), wsi.TargetCollection(), wsi.Note(keywordEnter, in))
	err := s.server.DeleteCollection(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetCollection(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) DeleteItem(in string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsDelete(), wsi.TargetItem(), wsi.Note(keywordEnter, in))
	err := s.server.DeleteItem(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetItem(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) DeleteVersion(in string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsDelete(), wsi.TargetVersion(), wsi.Note(keywordEnter, in))
	err := s.server.DeleteVersion(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetVersion(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) DeleteResource(in string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsDelete(), wsi.TargetResource(), wsi.Note(keywordEnter, in))
	err := s.server.DeleteResource(in)
	s.log(err, ts, wsi.IsDelete(), wsi.TargetResource(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) FindCollections(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsFind(), wsi.TargetCollection(), wsi.Note(keywordEnter, l, o))
	results, err := s.server.FindCollections(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetCollection(), wsi.Note(keywordExit, l, o, len(results)))
	return results, err
}

func (s *MiddlewareMonitor) FindItems(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Item, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsFind(), wsi.TargetItem(), wsi.Note(keywordEnter, l, o))
	results, err := s.server.FindItems(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetItem(), wsi.Note(keywordExit, l, o, len(results)))
	return results, err
}

func (s *MiddlewareMonitor) FindVersions(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Version, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsFind(), wsi.TargetVersion(), wsi.Note(keywordEnter, l, o))
	results, err := s.server.FindVersions(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetVersion(), wsi.Note(keywordExit, l, o, len(results)))
	return results, err
}

func (s *MiddlewareMonitor) FindResources(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsFind(), wsi.TargetResource(), wsi.Note(keywordEnter, l, o))
	results, err := s.server.FindResources(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetResource(), wsi.Note(keywordExit, l, o, len(results)))
	return results, err
}

func (s *MiddlewareMonitor) FindLinks(l int32, o int32, q []*wyc.QueryDesc) ([]*wyc.Link, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsFind(), wsi.TargetLink(), wsi.Note(keywordEnter, l, o))
	results, err := s.server.FindLinks(l, o, q)
	s.log(err, ts, wsi.IsFind(), wsi.TargetLink(), wsi.Note(keywordExit, l, o, len(results)))
	return results, err
}

func (s *MiddlewareMonitor) PublishedVersion(in string) (*wyc.Version, error) {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsPublished(), wsi.TargetVersion(), wsi.Note(keywordEnter, in))
	result, err := s.server.PublishedVersion(in)
	s.log(err, ts, wsi.IsPublished(), wsi.TargetVersion(), wsi.Note(keywordExit, in))
	return result, err
}

func (s *MiddlewareMonitor) SetPublishedVersion(in string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsPublish(), wsi.TargetVersion(), wsi.Note(keywordEnter, in))
	err := s.server.SetPublishedVersion(in)
	s.log(err, ts, wsi.IsPublish(), wsi.TargetVersion(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) UpdateVersionFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsUpdate(), wsi.TargetVersion(), wsi.Note(keywordEnter, in))
	err := s.server.UpdateVersionFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetVersion(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) UpdateItemFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(keywordEnter, in))
	err := s.server.UpdateItemFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) UpdateCollectionFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsUpdate(), wsi.TargetCollection(), wsi.Note(keywordEnter, in))
	err := s.server.UpdateCollectionFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetCollection(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) UpdateResourceFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsUpdate(), wsi.TargetResource(), wsi.Note(keywordEnter, in))
	err := s.server.UpdateResourceFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetResource(), wsi.Note(keywordExit, in))
	return err
}

func (s *MiddlewareMonitor) UpdateLinkFacets(in string, m map[string]string) error {
	ts := time.Now().UnixNano()
	s.log(nil, ts, wsi.IsUpdate(), wsi.TargetLink(), wsi.Note(keywordEnter, in))
	err := s.server.UpdateLinkFacets(in, m)
	s.log(err, ts, wsi.IsUpdate(), wsi.TargetLink(), wsi.Note(keywordExit, in))
	return err
}
