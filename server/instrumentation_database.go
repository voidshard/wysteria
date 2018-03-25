/*
Database instrumentation
 - This is an implementation of Database that accepts a logging monitor and another implementor
   of Database. All this does is pass any calls to the database it wraps & passes
   back the result(s). In the process, every action is logged to the given monitor.
*/

package main

import (
	wyc "github.com/voidshard/wysteria/common"
	wyd "github.com/voidshard/wysteria/server/database"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	"time"
)

func newDatabaseMonitor(db wyd.Database, monitor *wsi.Monitor) wyd.Database {
	return &DatabaseMonitor{database: db, monitor: monitor}
}

type DatabaseMonitor struct {
	database wyd.Database
	monitor  *wsi.Monitor
}

// Call on our monitor to do the actual logging business
//
func (s *DatabaseMonitor) log(err error, t int64, opts ...wsi.Opt) {
	area := "database"
	go func() {
		opts = append(opts, wsi.Time(time.Now().UnixNano()-t))
		if err != nil {
			s.monitor.Err(err, area, opts...)
			return
		}
		s.monitor.Log(area, opts...)
	}()
}

// Set given version id as published
func (d *DatabaseMonitor) SetPublished(in string) error {
	ts := time.Now().UnixNano()
	err := d.database.SetPublished(in)
	d.log(
		err, ts,
		wsi.IsPublish(), wsi.TargetVersion(),
		wsi.Note(in),
	)
	return err
}

// Given an Item ID, return the ID of the current PublishedVersion version (if any)
func (d *DatabaseMonitor) Published(in string) (*wyc.Version, error) {
	ts := time.Now().UnixNano()
	result, err := d.database.Published(in)
	d.log(
		err, ts,
		wsi.IsPublished(), wsi.TargetVersion(),
		wsi.Note(in),
	)
	return result, err
}

// Insert a collection into the db, return created Id
func (d *DatabaseMonitor) InsertCollection(in *wyc.Collection) (string, error) {
	ts := time.Now().UnixNano()
	result, err := d.database.InsertCollection(in)
	d.log(
		err, ts,
		wsi.IsCreate(), wsi.TargetCollection(),
		wsi.Note(in.Parent, in.Id, in.Name),
	)
	return result, err
}

// Insert an item into the db, return created Id
func (d *DatabaseMonitor) InsertItem(in *wyc.Item) (string, error) {
	ts := time.Now().UnixNano()
	result, err := d.database.InsertItem(in)
	d.log(err, ts,
		wsi.IsCreate(), wsi.TargetItem(),
		wsi.Note(in.Parent, in.Id, in.ItemType, in.Variant),
	)
	return result, err
}

// Insert a Version into the db, return created Id
func (d *DatabaseMonitor) InsertNextVersion(in *wyc.Version) (string, int32, error) {
	ts := time.Now().UnixNano()
	resultId, resultVer, err := d.database.InsertNextVersion(in)
	d.log(
		err, ts,
		wsi.IsCreate(), wsi.TargetVersion(),
		wsi.Note(in.Parent, in.Id, in.Number),
	)
	return resultId, resultVer, err
}

// Insert resource into the db, return created Id
func (d *DatabaseMonitor) InsertResource(in *wyc.Resource) (string, error) {
	ts := time.Now().UnixNano()
	result, err := d.database.InsertResource(in)
	d.log(err, ts, wsi.IsCreate(), wsi.TargetResource(), wsi.Note(in.Parent, in.Id, in.Name))
	return result, err
}

// Insert link into the db, return created Id
func (d *DatabaseMonitor) InsertLink(in *wyc.Link) (string, error) {
	ts := time.Now().UnixNano()
	result, err := d.database.InsertLink(in)
	d.log(err, ts, wsi.IsCreate(), wsi.TargetLink(), wsi.Note(in.Id, in.Src, in.Dst))
	return result, err
}

// Retrieve collections indicated by the given Id(s) from the db
func (d *DatabaseMonitor) RetrieveCollection(in ...string) ([]*wyc.Collection, error) {
	ts := time.Now().UnixNano()
	results, err := d.database.RetrieveCollection(in...)
	d.log(err, ts, wsi.IsFind(), wsi.TargetCollection(), wsi.Note(len(results)))
	return results, err
}

// Retrieve items indicated by the given Id(s) from the db
func (d *DatabaseMonitor) RetrieveItem(in ...string) ([]*wyc.Item, error) {
	ts := time.Now().UnixNano()
	results, err := d.database.RetrieveItem(in...)
	d.log(err, ts, wsi.IsFind(), wsi.TargetItem(), wsi.Note(len(results)))
	return results, err
}

// Retrieve versions indicated by the given Id(s) from the db
func (d *DatabaseMonitor) RetrieveVersion(in ...string) ([]*wyc.Version, error) {
	ts := time.Now().UnixNano()
	results, err := d.database.RetrieveVersion(in...)
	d.log(err, ts, wsi.IsFind(), wsi.TargetVersion(), wsi.Note(len(results)))
	return results, err
}

// Retrieve resources indicated by the given Id(s) from the db
func (d *DatabaseMonitor) RetrieveResource(in ...string) ([]*wyc.Resource, error) {
	ts := time.Now().UnixNano()
	results, err := d.database.RetrieveResource(in...)
	d.log(err, ts, wsi.IsFind(), wsi.TargetResource(), wsi.Note(len(results)))
	return results, err
}

// Retrieve links indicated by the given Id(s) from the db
func (d *DatabaseMonitor) RetrieveLink(in ...string) ([]*wyc.Link, error) {
	ts := time.Now().UnixNano()
	results, err := d.database.RetrieveLink(in...)
	d.log(err, ts, wsi.IsFind(), wsi.TargetLink(), wsi.Note(len(results)))
	return results, err
}

// Save the updated facets on the given version
func (d *DatabaseMonitor) UpdateItem(id string, in *wyc.Item) error {
	ts := time.Now().UnixNano()
	err := d.database.UpdateItem(id, in)
	d.log(err, ts, wsi.IsUpdate(), wsi.TargetItem(), wsi.Note(id))
	return err
}

// Save the updated facets on the given item
func (d *DatabaseMonitor) UpdateVersion(id string, in *wyc.Version) error {
	ts := time.Now().UnixNano()
	err := d.database.UpdateVersion(id, in)
	d.log(err, ts, wsi.IsUpdate(), wsi.TargetVersion(), wsi.Note(id))
	return err
}

// Save the updated facets on the given collection
func (d *DatabaseMonitor) UpdateCollection(id string, in *wyc.Collection) error {
	ts := time.Now().UnixNano()
	err := d.database.UpdateCollection(id, in)
	d.log(err, ts, wsi.IsUpdate(), wsi.TargetCollection(), wsi.Note(id))
	return err
}

// Save the updated facets on the given resource
func (d *DatabaseMonitor) UpdateResource(id string, in *wyc.Resource) error {
	ts := time.Now().UnixNano()
	err := d.database.UpdateResource(id, in)
	d.log(err, ts, wsi.IsUpdate(), wsi.TargetResource(), wsi.Note(id))
	return err
}

// Save the updated facets on the given link
func (d *DatabaseMonitor) UpdateLink(id string, in *wyc.Link) error {
	ts := time.Now().UnixNano()
	err := d.database.UpdateLink(id, in)
	d.log(err, ts, wsi.IsUpdate(), wsi.TargetLink(), wsi.Note(id))
	return err
}

// Delete collection(s) with the given Id(s)
func (d *DatabaseMonitor) DeleteCollection(in ...string) error {
	ts := time.Now().UnixNano()
	err := d.database.DeleteCollection(in...)
	d.log(err, ts, wsi.IsDelete(), wsi.TargetCollection(), wsi.Note(in))
	return err
}

// Delete item(s) with the given Id(s)
func (d *DatabaseMonitor) DeleteItem(in ...string) error {
	ts := time.Now().UnixNano()
	err := d.database.DeleteItem(in...)
	d.log(err, ts, wsi.IsDelete(), wsi.TargetItem(), wsi.Note(in))
	return err
}

// Delete version(s) with the given Id(s)
func (d *DatabaseMonitor) DeleteVersion(in ...string) error {
	ts := time.Now().UnixNano()
	err := d.database.DeleteVersion(in...)
	d.log(err, ts, wsi.IsDelete(), wsi.TargetVersion(), wsi.Note(in))
	return err
}

// Delete resource(s) with the given Id(s)
func (d *DatabaseMonitor) DeleteResource(in ...string) error {
	ts := time.Now().UnixNano()
	err := d.database.DeleteResource(in...)
	d.log(err, ts, wsi.IsDelete(), wsi.TargetResource(), wsi.Note(in))
	return err
}

// Delete link(s) with the given Id(s)
func (d *DatabaseMonitor) DeleteLink(in ...string) error {
	ts := time.Now().UnixNano()
	err := d.database.DeleteLink(in...)
	d.log(err, ts, wsi.IsDelete(), wsi.TargetLink(), wsi.Note(in))
	return err
}

// kill connection to db
func (d *DatabaseMonitor) Close() error {
	return d.database.Close()
}
