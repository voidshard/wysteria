/*
'main' server module.

The role of the server is simple really, we provide functions to the middleware to call when client requests arrive.
Our functions perform some sanity checks on given data we've got from the client, then we call the correct
searchbase / database functions and return what ever result we got, including errors.

We don't bother to encode or decode data here, that's the job of the layer above us.

Goals:
- Only objects with all required fields set may be created
- The database is the canonical source and should be updated / inserted into first, that way by the time things
  are searchable, they are already retrievable.
- The searchbase is where things around found in order to be retrieved, we should delete things from here first.
*/

package main

import (
	"errors"
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"strings"
)

// Main server struct
//  Here we implement the middleware/ServerHandler interface so we can hand a pointer to our server to the
//  middleware layer so the appropriate function calls can be routed through the logic here.
type WysteriaServer struct {
	GracefulShutdownTime time.Duration

	settings *configuration

	database          wdb.Database
	searchbase        wsb.Searchbase
	middleware_server wcm.EndpointServer
}

var (
	reservedItemFacets = []string{wyc.FacetCollection}
	reservedVerFacets = []string{wyc.FacetCollection, wyc.FacetItemType, wyc.FacetItemVariant}
)

// Update facets on the version with the given ID
//
func (s *WysteriaServer) UpdateVersionFacets(id string, update map[string]string) error {
	vers, err := s.database.RetrieveVersion(id)
	if err != nil {
		return err
	}
	if len(vers) != 1 { // there must be something to update
		return errors.New(fmt.Sprintf("No Version found with id %s", id))
	}

	version := vers[0]
	for key, value := range update {
		// prevent the updating of reserved keys
		if ListContains(strings.ToLower(key), reservedVerFacets) {
			continue
		}
		version.Facets[key] = value
	}

	err = s.database.UpdateVersion(id, version)
	if err != nil {
		return err
	}
	return s.searchbase.UpdateVersion(id, version)
}

// Update facets on the item with the given ID
//
func (s *WysteriaServer) UpdateItemFacets(id string, update map[string]string) error {
	vers, err := s.database.RetrieveItem(id)
	if err != nil {
		return err
	}
	if len(vers) != 1 { // there must be something to update
		return errors.New(fmt.Sprintf("No Item found with id %s", id))
	}

	item := vers[0]
	for key, value := range update {
		// prevent the updating of reserved keys
		if ListContains(strings.ToLower(key), reservedItemFacets) {
			continue
		}
		item.Facets[key] = value
	}

	err = s.database.UpdateItem(id, item)
	if err != nil {
		return err
	}
	return s.searchbase.UpdateItem(id, item)
}

// Create a collection with the given name.
//   Any ID set by the client is ignored.
//   Will fail if name is empty or a collection with the given name already exists
func (s *WysteriaServer) CreateCollection(name string) (string, error) {
	if name == "" { // Check required field
		return "", errors.New("Name required for Collection")
	}

	id := NewId()
	obj := &wyc.Collection{Name: name, Id: id}

	err := s.database.InsertCollection(id, obj)
	if err != nil {
		return "", err
	}
	return id, s.searchbase.InsertCollection(id, obj)
}

// Create a item based on the given item.
//   We use the given item type, variant & facets.
//   Any ID set by the client is ignored.
//   Will fail if the parent collection already has a item with the given type & variant.
//   Will fail is parent id, type or variant are empty.
func (s *WysteriaServer) CreateItem(in *wyc.Item) (string, error) {
	if in.Parent == "" || in.ItemType == "" || in.Variant == "" {
		return "", errors.New("Require Parent, ItemType, Variant to be set")
	}

	_, ok := in.Facets[wyc.FacetCollection]
	if !ok {
		return "", errors.New(fmt.Sprintf("Required facet %s not set", wyc.FacetCollection))
	}

	in.Facets[wyc.FacetItemType] = in.ItemType
	in.Facets[wyc.FacetItemVariant] = in.Variant

	in.Id = NewId()
	err := s.database.InsertItem(in.Id, in)
	if err != nil {
		return "", err
	}
	return in.Id, s.searchbase.InsertItem(in.Id, in)
}

// Create a version based on the given item.
//   We use the given parent value.
//   Any ID set by the client is ignored.
//   Will fail if one of our required facets (collection, item type, item variant) isn't set.
func (s *WysteriaServer) CreateVersion(in *wyc.Version) (string, int32, error) {
	if in.Parent == "" {
		return "", 0, errors.New("Require Parent to be set")
	}

	for _, facet_key := range reservedVerFacets {
		_, ok := in.Facets[facet_key]
		if !ok {
			return "", 0, errors.New(fmt.Sprintf("Required facet '%s' not set", facet_key))
		}
	}

	in.Id = NewId()
	version_number, err := s.database.InsertNextVersion(in.Id, in)
	if err != nil {
		return "", 0, err
	}

	in.Number = version_number
	return in.Id, in.Number, s.searchbase.InsertVersion(in.Id, in)
}

// Create resource with the given base settings.
// Any ID set by client will be ignored.
// Will fail if the parent or location values aren't set.
// Note we don't enforce the use of a name or resource type .. but it's recommended to use them.
func (s *WysteriaServer) CreateResource(in *wyc.Resource) (string, error) {
	if in.Parent == "" || in.Location == "" {
		return "", errors.New("Require Parent, Location to be set")
	}

	in.Id = NewId()
	err := s.database.InsertResource(in.Id, in)
	if err != nil {
		return "", err
	}
	return in.Id, s.searchbase.InsertResource(in.Id, in)
}

// Create link with the given base settings.
// Any ID set by client will be ignored.
// Will fail if the source or destination fields aren't set, or if they're the same.
// Note we don't enforce the use of a link name, but it's recommended to use one.
func (s *WysteriaServer) CreateLink(in *wyc.Link) (string, error) {
	if in.Src == "" || in.Dst == "" {
		return "", errors.New("Require Src, Dst to be set")
	}

	// Not a perfect check but hopefully no one tries this too hard.
	// It shouldn't break anything it's just ... pointless.
	if in.Src == in.Dst {
		return "", errors.New("You may not link something to itself")
	}

	// We're good to create our link
	in.Id = NewId()
	err := s.database.InsertLink(in.Id, in)
	if err != nil {
		return "", err
	}
	return in.Id, s.searchbase.InsertLink(in.Id, in)
}

// Given some ids, build the appropriate query to return all of their children
func childrenOf(ids ...string) []*wyc.QueryDesc {
	result := []*wyc.QueryDesc{}
	for _, id := range ids {
		result = append(result, &wyc.QueryDesc{Parent: id})
	}
	return result
}

// Delete some collection from the system.
// Assuming this works, we kick off a routine to kill all of the children.
// Please be aware that delete operations, especially of collections, are heavy operations that introduce a number
// of race conditions for people still using the collection (or children of it).
func (s *WysteriaServer) DeleteCollection(id string) error {
	err := s.searchbase.DeleteCollection(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteCollection(id)
	if err != nil {
		return err
	}

	go func() {
		// Kick off a routine to slay the children
		children, err := s.searchbase.QueryItem(0, 0, childrenOf(id)...)
		if err == nil {
			for _, child := range children {
				s.DeleteItem(child)
			}
		}
	}()
	return nil
}

// Given some ids, build query to find all links mentioning those ids (as either src or dst)
func linkedTo(ids ...string) []*wyc.QueryDesc {
	result := []*wyc.QueryDesc{}
	for _, id := range ids {
		result = append(
			result,
			&wyc.QueryDesc{LinkSrc: id},
			&wyc.QueryDesc{LinkDst: id},
		)
	}
	return result
}

// Delete some item from the system.
// Assuming this works, we kick off a routine to kill all of the children.
func (s *WysteriaServer) DeleteItem(id string) error {
	err := s.searchbase.DeleteItem(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteItem(id)
	if err != nil {
		return err
	}

	go func() {
		// kick off a routine to kill links that mention this
		linked, err := s.searchbase.QueryLink(0, 0, linkedTo(id)...)
		if err == nil {
			s.searchbase.DeleteLink(linked...)
			s.database.DeleteLink(linked...)
		}
	}()

	go func() {
		// Kick off a routine to slay children
		children, err := s.searchbase.QueryVersion(0, 0, childrenOf(id)...)
		if err == nil {
			for _, child := range children {
				s.DeleteVersion(child)
			}
		}
	}()
	return nil
}

// Delete some version from the system.
// Assuming this works, we kick off a routine to kill all of the children.
func (s *WysteriaServer) DeleteVersion(id string) error {
	err := s.searchbase.DeleteVersion(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteVersion(id)
	if err != nil {
		return err
	}

	go func() {
		// kick off a routine to kill links that mention this
		linked, err := s.searchbase.QueryLink(0, 0, linkedTo(id)...)
		if err == nil {
			s.searchbase.DeleteLink(linked...)
			s.database.DeleteLink(linked...)
		}
	}()

	go func() {
		// Kick off a routine to slay children
		children, err := s.searchbase.QueryResource(0, 0, childrenOf(id)...)
		if err == nil {
			for _, child := range children {
				s.DeleteResource(child)
			}
		}
	}()
	return nil
}

// Delete some resource from the system.
func (s *WysteriaServer) DeleteResource(id string) error {
	err := s.searchbase.DeleteResource(id)
	if err != nil {
		return err
	}

	return s.database.DeleteResource(id)
}

func (s *WysteriaServer) FindCollections(qs []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	ids, err := s.searchbase.QueryCollection(0, 0, qs...)
	if err != nil {
		return nil, err
	}
	return s.database.RetrieveCollection(ids...)
}

func (s *WysteriaServer) FindItems(qs []*wyc.QueryDesc) ([]*wyc.Item, error) {
	ids, err := s.searchbase.QueryItem(0, 0, qs...)
	if err != nil {
		return nil, err
	}
	return s.database.RetrieveItem(ids...)
}

func (s *WysteriaServer) FindVersions(qs []*wyc.QueryDesc) ([]*wyc.Version, error) {
	ids, err := s.searchbase.QueryVersion(0, 0, qs...)
	if err != nil {
		return nil, err
	}
	return s.database.RetrieveVersion(ids...)
}

func (s *WysteriaServer) FindResources(qs []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	ids, err := s.searchbase.QueryResource(0, 0, qs...)
	if err != nil {
		return nil, err
	}
	return s.database.RetrieveResource(ids...)
}

func (s *WysteriaServer) FindLinks(qs []*wyc.QueryDesc) ([]*wyc.Link, error) {
	ids, err := s.searchbase.QueryLink(0, 0, qs...)
	if err != nil {
		return nil, err
	}
	return s.database.RetrieveLink(ids...)
}

func (s *WysteriaServer) PublishedVersion(item_id string) (*wyc.Version, error) {
	return s.database.Published(item_id)
}

func (s *WysteriaServer) SetPublishedVersion(version_id string) error {
	return s.database.SetPublished(version_id)
}

func (s *WysteriaServer) Shutdown() {
	go s.closeConnection() // send a routine to kill off connections nicely

	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	msg := "Shutdown request recieved, giving %s for connections to close gracefully"
	log.Println(fmt.Sprintf(msg, s.GracefulShutdownTime))

	select {
	case <-time.After(s.GracefulShutdownTime * 2):
		return
	case s := <-ch:
		log.Fatalf("Received signal %s: terminating immediately", s)
	}
}

func (s *WysteriaServer) closeConnection() {
	go s.middleware_server.Shutdown()

	select {
	case <-time.After(s.GracefulShutdownTime):
		break
	}
	s.database.Close()
	s.searchbase.Close()
}

func (s *WysteriaServer) Run() error {
	msg := "Opening %s %s %s:%d"

	// [1] Connect / spin up the database
	log.Println(fmt.Sprintf(msg, "database", s.settings.Database.Driver, s.settings.Database.Host, s.settings.Database.Port))
	database, err := wdb.Connect(&s.settings.Database)
	if err != nil {
		return err
	}
	s.database = database

	// [2] Connect / spin up the searchbase
	log.Println(fmt.Sprintf(msg, "searchbase", s.settings.Searchbase.Driver, s.settings.Searchbase.Host, s.settings.Searchbase.Port))
	searchbase, err := wsb.Connect(&s.settings.Searchbase)
	if err != nil {
		return err
	}
	s.searchbase = searchbase

	// [3] Lastly, spin up or connect to whatever is bring us requests
	log.Println(fmt.Sprintf("Initializing middleware %s", s.settings.Middleware.Driver))
	mware_server, err := wcm.NewServer(s.settings.Middleware.Driver)
	if err != nil {
		return err
	}
	s.middleware_server = mware_server

	log.Println("Spinning up middleware & waiting for connections")
	return mware_server.ListenAndServe(s.settings.Middleware.Config, s)
}
