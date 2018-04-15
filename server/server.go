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
	"fmt"
	wyc "github.com/voidshard/wysteria/common"
	wcm "github.com/voidshard/wysteria/common/middleware"
	wdb "github.com/voidshard/wysteria/server/database"
	wsi "github.com/voidshard/wysteria/server/instrumentation"
	wsb "github.com/voidshard/wysteria/server/searchbase"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Main server struct
//  Here we implement the middleware/ServerHandler interface so we can hand a pointer to our server to the
//  middleware layer so the appropriate function calls can be routed through the logic here.
type WysteriaServer struct {
	GracefulShutdownTime time.Duration
	refuseClientRequests bool
	refuseClientReason   string
	refuseClientLock     sync.RWMutex

	settings *configuration

	database         wdb.Database
	searchbase       wsb.Searchbase
	middlewareServer wcm.EndpointServer
	monitor          *wsi.Monitor
}

const (
	// default used internally when making otherwise limitless queries
	defaultQueryLimit = 10000
)

var (
	reservedColFacets  = []string{wyc.FacetCollection}
	reservedItemFacets = []string{wyc.FacetCollection}
	reservedVerFacets  = []string{wyc.FacetCollection, wyc.FacetItemType, wyc.FacetItemVariant}
)

// Update facets on the collection with the given ID
//
func (s *WysteriaServer) UpdateCollectionFacets(id string, update map[string]string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	if update == nil {
		return nil
	}

	results, err := s.database.RetrieveCollection(id)
	if err != nil {
		return err
	}
	if len(results) != 1 {
		return fmt.Errorf("%s: no Collection found with id %s", wyc.ErrorNotFound, id)
	}

	for key, value := range update {
		if ListContains(strings.ToLower(key), reservedColFacets) {
			continue
		}
		results[0].Facets[key] = value
	}

	err = s.database.UpdateCollection(id, results[0])
	if err != nil {
		return err
	}
	return s.searchbase.UpdateCollection(id, results[0])
}

// Update facets on the resource with the given ID
//
func (s *WysteriaServer) UpdateResourceFacets(id string, update map[string]string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	if update == nil {
		return nil
	}

	results, err := s.database.RetrieveResource(id)
	if err != nil {
		return err
	}
	if len(results) != 1 {
		return fmt.Errorf("%s:no Resource found with id %s", wyc.ErrorNotFound, id)
	}

	for key, value := range update {
		results[0].Facets[key] = value
	}

	err = s.database.UpdateResource(id, results[0])
	if err != nil {
		return err
	}
	return s.searchbase.UpdateResource(id, results[0])
}

// Update facets on the link with the given ID
//
func (s *WysteriaServer) UpdateLinkFacets(id string, update map[string]string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	if update == nil {
		return nil
	}

	results, err := s.database.RetrieveLink(id)
	if err != nil {
		return err
	}
	if len(results) != 1 {
		return fmt.Errorf("%s: no Link found with id %s", wyc.ErrorNotFound, id)
	}

	for key, value := range update {
		results[0].Facets[key] = value
	}

	err = s.database.UpdateLink(id, results[0])
	if err != nil {
		return err
	}
	return s.searchbase.UpdateLink(id, results[0])
}

// Update facets on the version with the given ID
//
func (s *WysteriaServer) UpdateVersionFacets(id string, update map[string]string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	if update == nil {
		return nil
	}

	vers, err := s.database.RetrieveVersion(id)
	if err != nil {
		return err
	}
	if len(vers) != 1 { // there must be something to update
		return fmt.Errorf("%s: no Version found with id %s", wyc.ErrorNotFound, id)
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
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	if update == nil {
		return nil
	}

	vers, err := s.database.RetrieveItem(id)
	if err != nil {
		return err
	}
	if len(vers) != 1 { // there must be something to update
		return fmt.Errorf("%s: no Item found with id %s", wyc.ErrorNotFound, id)
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
func (s *WysteriaServer) CreateCollection(in *wyc.Collection) (string, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return "", err
	}

	if in.Name == "" { // Check required field
		return "", fmt.Errorf("%s: name required for Collection", wyc.ErrorNotFound)
	}
	if in.Facets == nil {
		in.Facets = make(map[string]string)
	}

	// set the parent name
	if in.Parent == "" {
		in.Facets[wyc.FacetCollection] = wyc.FacetRootCollection
	} else {
		parent, err := s.database.RetrieveCollection(in.Parent)
		if err != nil {
			return "", err
		}
		if len(parent) != 1 {
			return "", fmt.Errorf("%s: unable to find parent with id %s", wyc.ErrorNotFound, in.Parent)
		}
		in.Facets[wyc.FacetCollection] = parent[0].Name
	}

	id, err := s.database.InsertCollection(in)
	if err != nil {
		return "", err
	}
	return id, s.searchbase.InsertCollection(id, in)
}

// Create a item based on the given item.
//   We use the given item type, variant & facets.
//   Any ID set by the client is ignored.
//   Will fail if the parent collection already has a item with the given type & variant.
//   Will fail is parent id, type or variant are empty.
func (s *WysteriaServer) CreateItem(in *wyc.Item) (string, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return "", err
	}

	if in.Parent == "" || in.ItemType == "" || in.Variant == "" {
		return "", fmt.Errorf("%s: require Parent, ItemType, Variant to be set", wyc.ErrorInvalid)
	}
	if in.Facets == nil {
		in.Facets = make(map[string]string)
	}

	_, ok := in.Facets[wyc.FacetCollection]
	if !ok {
		return "", fmt.Errorf("%s: required facet %s not set", wyc.ErrorInvalid, wyc.FacetCollection)
	}

	in.Facets[wyc.FacetItemType] = in.ItemType
	in.Facets[wyc.FacetItemVariant] = in.Variant

	id, err := s.database.InsertItem(in)
	if err != nil {
		return "", err
	}
	in.Id = id
	return in.Id, s.searchbase.InsertItem(in.Id, in)
}

// Create a version based on the given item.
//   We use the given parent value.
//   Any ID set by the client is ignored.
//   Will fail if one of our required facets (collection, item type, item variant) isn't set.
func (s *WysteriaServer) CreateVersion(in *wyc.Version) (string, int32, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return "", 0, err
	}

	if in.Parent == "" {
		return "", 0, fmt.Errorf("%s: require Parent to be set", wyc.ErrorInvalid)
	}
	if in.Facets == nil {
		in.Facets = make(map[string]string)
	}

	for _, facet_key := range reservedVerFacets {
		_, ok := in.Facets[facet_key]
		if !ok {
			return "", 0, fmt.Errorf("%s: required facet '%s' not set", wyc.ErrorInvalid, facet_key)
		}
	}

	id, versionNumber, err := s.database.InsertNextVersion(in)
	if err != nil {
		return "", 0, err
	}
	in.Id = id
	in.Number = versionNumber
	return in.Id, in.Number, s.searchbase.InsertVersion(in.Id, in)
}

// Create resource with the given base settings.
// Any ID set by client will be ignored.
// Will fail if the parent or location values aren't set.
// Note we don't enforce the use of a name or resource type .. but it's recommended to use them.
func (s *WysteriaServer) CreateResource(in *wyc.Resource) (string, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return "", err
	}

	if in.Parent == "" || in.Location == "" {
		return "", fmt.Errorf("%s: Require Parent, Location to be set", wyc.ErrorInvalid)
	}
	if in.Facets == nil {
		in.Facets = make(map[string]string)
	}

	id, err := s.database.InsertResource(in)
	if err != nil {
		return "", err
	}
	in.Id = id
	return in.Id, s.searchbase.InsertResource(in.Id, in)
}

// Create link with the given base settings.
// Any ID set by client will be ignored.
// Will fail if the source or destination fields aren't set, or if they're the same.
// Note we don't enforce the use of a link name, but it's recommended to use one.
func (s *WysteriaServer) CreateLink(in *wyc.Link) (string, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return "", err
	}

	if in.Src == "" || in.Dst == "" {
		return "", fmt.Errorf("%s: Require Src, Dst to be set", wyc.ErrorInvalid)
	}

	// Not a perfect check but hopefully no one tries this too hard.
	// It shouldn't break anything it's just ... pointless.
	if in.Src == in.Dst {
		return "", fmt.Errorf("%s: You may not link something to itself", wyc.ErrorIllegal)
	}
	if in.Facets == nil {
		in.Facets = make(map[string]string)
	}

	// We're good to create our link
	id, err := s.database.InsertLink(in)
	if err != nil {
		return "", err
	}
	in.Id = id
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
func (s *WysteriaServer) DeleteCollection(id string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}

	childCollections, err := s.searchbase.QueryCollection(defaultQueryLimit, 0, childrenOf(id)...)
	if err != nil {
		return err
	}
	if len(childCollections) > 0 {
		return fmt.Errorf("%s: unable to delete: There are %d child collections", wyc.ErrorIllegal, len(childCollections))
	}

	children, err := s.searchbase.QueryItem(defaultQueryLimit, 0, childrenOf(id)...)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return fmt.Errorf("%s: unable to delete: There are %d child items", wyc.ErrorIllegal, len(children))
	}

	err = s.searchbase.DeleteCollection(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteCollection(id)
	if err != nil {
		return err
	}

	return err
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
func (s *WysteriaServer) DeleteItem(id string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}

	children, err := s.searchbase.QueryVersion(defaultQueryLimit, 0, childrenOf(id)...)
	if err != nil {
		return err
	}
	if len(children) > 0 {
		return fmt.Errorf("%s: unable to delete: There are %d child versions", wyc.ErrorIllegal, len(children))
	}

	err = s.searchbase.DeleteItem(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteItem(id)
	if err != nil {
		return err
	}

	linked, err := s.searchbase.QueryLink(defaultQueryLimit, 0, linkedTo(id)...)
	if err == nil && len(linked) > 0 {
		s.searchbase.DeleteLink(linked...)
		s.database.DeleteLink(linked...)
	}

	return nil
}

// Delete some version from the system.
func (s *WysteriaServer) DeleteVersion(id string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}

	err = s.searchbase.DeleteVersion(id)
	if err != nil {
		return err
	}

	err = s.database.DeleteVersion(id)
	if err != nil {
		return err
	}

	linked, err := s.searchbase.QueryLink(defaultQueryLimit, 0, linkedTo(id)...)
	if err == nil && len(linked) > 0 {
		s.searchbase.DeleteLink(linked...)
		s.database.DeleteLink(linked...)
	}

	children, err := s.searchbase.QueryResource(defaultQueryLimit, 0, childrenOf(id)...)
	if err == nil && len(children) > 0 {
		s.searchbase.DeleteResource(children...)
		s.database.DeleteResource(children...)
	}

	return nil
}

// Delete some resource from the system.
func (s *WysteriaServer) DeleteResource(id string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}

	err = s.searchbase.DeleteResource(id)
	if err != nil {
		return err
	}

	return s.database.DeleteResource(id)
}

// Use searchbase to perform search, return any matching collections from database
func (s *WysteriaServer) FindCollections(limit, offset int32, qs []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}

	ids, queries := ExtractIdQueries(qs)
	if len(queries) > 0 || len(ids) == 0 {
		found, err := s.searchbase.QueryCollection(int(limit), int(offset), queries...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}

	if len(ids) < 1 {
		return nil, nil
	}
	return s.database.RetrieveCollection(ids...)
}

// Use searchbase to perform search, return any matching items from database
func (s *WysteriaServer) FindItems(limit, offset int32, qs []*wyc.QueryDesc) ([]*wyc.Item, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}

	ids, queries := ExtractIdQueries(qs)
	if len(queries) > 0 || len(ids) == 0 {
		found, err := s.searchbase.QueryItem(int(limit), int(offset), queries...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}

	if len(ids) < 1 {
		return nil, nil
	}
	return s.database.RetrieveItem(ids...)
}

// Use searchbase to perform search, return any matching versions from database
func (s *WysteriaServer) FindVersions(limit, offset int32, qs []*wyc.QueryDesc) ([]*wyc.Version, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}

	ids, queries := ExtractIdQueries(qs)
	if len(queries) > 0 || len(ids) == 0 {
		found, err := s.searchbase.QueryVersion(int(limit), int(offset), queries...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}

	if len(ids) < 1 {
		return nil, nil
	}
	return s.database.RetrieveVersion(ids...)
}

// Use searchbase to perform search, return any matching resources from database
func (s *WysteriaServer) FindResources(limit, offset int32, qs []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}

	ids, queries := ExtractIdQueries(qs)
	if len(queries) > 0 || len(ids) == 0 {
		found, err := s.searchbase.QueryResource(int(limit), int(offset), queries...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}

	if len(ids) < 1 {
		return nil, nil
	}
	return s.database.RetrieveResource(ids...)
}

// Use searchbase to perform search, return any matching links from database
func (s *WysteriaServer) FindLinks(limit, offset int32, qs []*wyc.QueryDesc) ([]*wyc.Link, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}

	ids, queries := ExtractIdQueries(qs)
	if len(queries) > 0 || len(ids) == 0 {
		found, err := s.searchbase.QueryLink(int(limit), int(offset), queries...)
		if err != nil {
			return nil, err
		}
		ids = append(ids, found...)
	}

	if len(ids) < 1 {
		return nil, nil
	}
	return s.database.RetrieveLink(ids...)
}

// Get the published version whose parent item's id is the given item id
func (s *WysteriaServer) PublishedVersion(item_id string) (*wyc.Version, error) {
	err := s.shouldServeRequest()
	if err != nil {
		return nil, err
	}
	return s.database.Published(item_id)
}

// Get the version with the given id as published
func (s *WysteriaServer) SetPublishedVersion(version_id string) error {
	err := s.shouldServeRequest()
	if err != nil {
		return err
	}
	return s.database.SetPublished(version_id)
}

// Shutdown the main server
func (s *WysteriaServer) Shutdown() {
	s.monitor.Warn("shutdown", wsi.InFunc("Shutdown()"))
	s.middlewareServer.Shutdown()
	s.database.Close()
	s.searchbase.Close()
}

// Set if the server should serve a client request.
func (s *WysteriaServer) setRefuseClientRequests(value bool, reason string) {
	s.monitor.Warn("refuseClients", wsi.InFunc("setRefuseClientRequests()"), wsi.Note(reason, value))

	s.refuseClientLock.Lock()
	defer s.refuseClientLock.Unlock()

	s.refuseClientRequests = value
	s.refuseClientReason = reason
}

// Returns an error if the server is set to not serve a request, and returns the set reason as an error.
func (s *WysteriaServer) shouldServeRequest() error {
	s.refuseClientLock.RLock()
	defer s.refuseClientLock.RUnlock()

	if s.refuseClientRequests {
		return fmt.Errorf("%s: %s", wyc.ErrorNotServing, s.refuseClientReason)
	}
	return nil
}

// Func to handle setting up of signal catcher
func (s *WysteriaServer) awaitSignal() {
	// ToDo: This func doesn't seem to always catch the signal, or perhaps the OS kills it before it can run?
	ch := make(chan os.Signal, 2)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGABRT)

	sig := <-ch // Wait for signal to be received from OS

	msg := fmt.Sprintf("Recieved signal: %s initiating shutdown", sig.String())
	log.Println(msg)
	s.setRefuseClientRequests(true, msg)

	select {
	case <-time.After(s.GracefulShutdownTime):
		break
	case sig := <-ch:
		log.Println("Recieved signal:", sig.String(), "shutting down")
		break
	}

	s.Shutdown()
	os.Exit(0)
}

// Setup all the logger(s) we've been asked to.
//  Wont fail a particular logger doesn't work - we will stop if all fail though.
//
func (s *WysteriaServer) setupMonitor() error {
	targets := []wsi.MonitorOutput{}
	for _, settings := range s.settings.Instrumentation {
		output, err := wsi.Connect(settings)
		if err != nil {
			log.Println(err)
			continue
		}
		targets = append(targets, output)
	}

	mon, err := wsi.NewMonitor(targets...)
	if err != nil {
		return err
	}

	s.monitor = mon
	return nil
}

// Start up the server, connect to any require remote service(s)
func (s *WysteriaServer) Run() error {
	go s.awaitSignal()
	msg := "Opening %s %s %s:%d"

	// [0] Start up the monitor
	log.Println("Initializing monitoring ...")
	err := s.setupMonitor()
	if err != nil {
		return err
	}
	err = s.monitor.Start(&s.settings.Health)
	if err != nil {
		return err
	}

	// [1] Connect / spin up the database
	log.Println(fmt.Sprintf(msg, "database", s.settings.Database.Driver, s.settings.Database.Host, s.settings.Database.Port))
	database, err := wdb.Connect(&s.settings.Database)
	if err != nil {
		return err
	}
	dblogger := newDatabaseMonitor(database, s.monitor)
	s.database = dblogger

	// [2] Connect / spin up the searchbase
	log.Println(fmt.Sprintf(msg, "searchbase", s.settings.Searchbase.Driver, s.settings.Searchbase.Host, s.settings.Searchbase.Port))
	searchbase, err := wsb.Connect(&s.settings.Searchbase)
	if err != nil {
		return err
	}
	sblogger := newSearchbaseMonitor(searchbase, s.monitor)
	s.searchbase = sblogger

	// [4] Spin up or connect to whatever is bring us requests
	log.Println(fmt.Sprintf("Initializing middleware %s", s.settings.Middleware.Driver))
	mwareServer, err := wcm.NewServer(s.settings.Middleware.Driver)
	if err != nil {
		return err
	}
	s.middlewareServer = mwareServer

	log.Println("[Booting]")

	shim := newMiddlewareMonitor(s.middlewareServer, s.monitor)
	return shim.ListenAndServe(&s.settings.Middleware, s)
}
