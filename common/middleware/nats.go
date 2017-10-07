/*
Implements the EndpointClient and EndpointServer for the Nats.io protocol.
*/

package middleware

import (
	"errors"
	"fmt"
	natsd "github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats"
	wyc "github.com/voidshard/wysteria/common"
	codec "github.com/voidshard/wysteria/common/middleware/ffjson_structs"
	"log"
	"time"
)

const (
	// Default nats settings
	natsDefaultHost = "localhost"
	natsDefaultPort = 4222
	natsQueueServer = "server_queue"

	// These routes indicate who is sending the message
	natsRouteServer   = "w.server."   // From a wysteria server
	natsRouteClient   = "w.client."   // From a client
	natsRouteInternal = "w.internal." // From the admin(s)

	// messages suffixes
	//  - these are used to indicate which function a client is hoping to call
	callCreateCollection = "cc"
	callCreateItem       = "ci"
	callCreateVersion    = "cv"
	callCreateResource   = "cr"
	callCreateLink       = "cl"

	callDeleteCollection = "dc"
	callDeleteItem       = "di"
	callDeleteVersion    = "dv"
	callDeleteResource   = "dr"

	callFindCollection = "fc"
	callFindItem       = "fi"
	callFindVersion    = "fv"
	callFindResource   = "fr"
	callFindLink       = "fl"

	callGetPublished = "gp"
	callSetPublished = "sp"

	callUpdateCollection = "uc"
	callUpdateVersion    = "uv"
	callUpdateItem       = "ui"
	callUpdateResource   = "ur"
	callUpdateLink       = "ul"

	// The server carves this number of chars off the end of the route
	// message to determine which func is being called by the suffix (above)
	callSuffixLength = 2
)

var (
	timeout = time.Second * 30
)

// Wrapper around nats client connection to provide our EndpointClient functions
type natsClient struct {
	conn *nats.Conn
}

// Create a new nats client connection wrapper
func newNatsClient() EndpointClient {
	return &natsClient{}
}

// Connect to server given the url
//  Url required by nats.io (from the docs)
//    nats://derek:pass@localhost:4222
//  As in,
//    nats://user:password@host:port
func (c *natsClient) Connect(config *Settings) error {
	if config.Config == "" {
		config.Config = fmt.Sprintf("nats://%s:%d", natsDefaultHost, natsDefaultPort)
	}

	raw, err := connect(config)
	if err != nil {
		return err
	}
	c.conn = raw
	return err
}

// Send some raw request to the server and await reply
func (c *natsClient) serverRequest(subject string, data []byte) (*nats.Msg, error) {
	return c.conn.Request(natsRouteClient+subject, data, timeout)
}

// Flush and close connection(s) to server
func (c *natsClient) Close() error {
	c.conn.Flush()
	c.conn.Close()
	return nil
}

// Util func to convert some string into an error.
// Basic stuff, just saves on typing in this file ..
func stringError(in string) error {
	if in == "" {
		return nil
	}
	return errors.New(in)
}

// Create new collection with given name, return new Id
//   - Collection name is required to be unique among collections
func (c *natsClient) CreateCollection(in *wyc.Collection) (id string, err error) {
	req := &codec.CreateReqCollection{Collection: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callCreateCollection, data)
	if err != nil {
		return
	}

	rep := &codec.CreateReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	return rep.Id, stringError(rep.Error)
}

// Send item creation request using the given item as a base, return new Id
// Required to include non empty fields for
//   - parent (collection Id)
//   - item type
//   - item variant
// Item facets are required to include
//   - grandparent collection name
func (c *natsClient) CreateItem(in *wyc.Item) (id string, err error) {
	req := &codec.CreateReqItem{Item: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callCreateItem, data)
	if err != nil {
		return
	}

	rep := &codec.CreateReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	return rep.Id, stringError(rep.Error)
}

// Send version creation request using given version as base, return new Id & new version number
// Required to include non empty fields for
//   - parent (item Id)
// Version facets are required to include
//   - grandparent collection name
//   - parent item type
//   - parent item variant
func (c *natsClient) CreateVersion(in *wyc.Version) (id string, num int32, err error) {
	req := &codec.CreateReqVersion{Version: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callCreateVersion, data)
	if err != nil {
		return
	}

	rep := &codec.CreateReplyVersion{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	return rep.Id, rep.Version, stringError(rep.Error)
}

// Send resource creation request using given resource as a base, return new Id
// Required to include non empty fields for
//   - parent (version Id)
func (c *natsClient) CreateResource(in *wyc.Resource) (id string, err error) {
	req := &codec.CreateReqResource{Resource: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callCreateResource, data)
	if err != nil {
		return
	}

	rep := &codec.CreateReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	return rep.Id, stringError(rep.Error)
}

// Send link creation request using given link as a base, return new Id
// Required to include non empty fields for
//   - source Id
//   - destination Id
func (c *natsClient) CreateLink(in *wyc.Link) (id string, err error) {
	req := &codec.CreateReqLink{Link: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callCreateLink, data)
	if err != nil {
		return
	}

	rep := &codec.CreateReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	return rep.Id, stringError(rep.Error)
}

// Util func to send a delete request, parse reply and return error
func (c *natsClient) genericDelete(id, subject string) error {
	req := &codec.DeleteReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.serverRequest(subject, data)
	if err != nil {
		return err
	}

	rep := &codec.DeleteReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return err
	}

	return stringError(rep.Error)
}

// Given the Id, delete some collection
func (c *natsClient) DeleteCollection(id string) error {
	return c.genericDelete(id, callDeleteCollection)
}

// Given the Id, delete some item
func (c *natsClient) DeleteItem(id string) error {
	return c.genericDelete(id, callDeleteItem)
}

// Given the Id, delete some version
func (c *natsClient) DeleteVersion(id string) error {
	return c.genericDelete(id, callDeleteVersion)
}

// Given the Id, delete some resource
func (c *natsClient) DeleteResource(id string) error {
	return c.genericDelete(id, callDeleteResource)
}

// Util to convert native wysteria QueryDesc list to a 'FindReq' that this protocol
// will send over the wire.
func toFindReq(limit, offset int32, query []*wyc.QueryDesc) *codec.FindReq {
	req := &codec.FindReq{
		Query:  []wyc.QueryDesc{},
		Limit:  limit,
		Offset: offset,
	}
	for _, q := range query {
		req.Query = append(req.Query, *q)
	}
	return req
}

// Given some list of QueryDescriptions, return matching collections
func (c *natsClient) FindCollections(limit, offset int32, query []*wyc.QueryDesc) (result []*wyc.Collection, err error) {
	data, err := toFindReq(limit, offset, query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callFindCollection, data)
	if err != nil {
		return
	}

	rep := &codec.FindReplyCollection{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	for _, r := range rep.All {
		tmp := r
		result = append(result, &tmp)
	}
	return result, stringError(rep.Error)
}

// Given some list of QueryDescriptions, return matching items
func (c *natsClient) FindItems(limit, offset int32, query []*wyc.QueryDesc) (result []*wyc.Item, err error) {
	data, err := toFindReq(limit, offset, query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callFindItem, data)
	if err != nil {
		return
	}

	rep := &codec.FindReplyItem{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	for _, r := range rep.All {
		tmp := r
		result = append(result, &tmp)
	}
	return result, stringError(rep.Error)
}

// Given some list of QueryDescriptions, return matching versions
func (c *natsClient) FindVersions(limit, offset int32, query []*wyc.QueryDesc) (result []*wyc.Version, err error) {
	data, err := toFindReq(limit, offset, query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callFindVersion, data)
	if err != nil {
		return
	}

	rep := &codec.FindReplyVersion{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	for _, r := range rep.All {
		tmp := r
		result = append(result, &tmp)
	}
	return result, stringError(rep.Error)
}

// Given some list of QueryDescriptions, return matching resources
func (c *natsClient) FindResources(limit, offset int32, query []*wyc.QueryDesc) (result []*wyc.Resource, err error) {
	data, err := toFindReq(limit, offset, query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callFindResource, data)
	if err != nil {
		return
	}

	rep := &codec.FindReplyResource{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	for _, r := range rep.All {
		tmp := r
		result = append(result, &tmp)
	}
	return result, stringError(rep.Error)
}

// Given some list of QueryDescriptions, return matching links
func (c *natsClient) FindLinks(limit, offset int32, query []*wyc.QueryDesc) (result []*wyc.Link, err error) {
	data, err := toFindReq(limit, offset, query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.serverRequest(callFindLink, data)
	if err != nil {
		return
	}

	rep := &codec.FindReplyLink{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return
	}

	for _, r := range rep.All {
		tmp := r
		result = append(result, &tmp)
	}
	return result, stringError(rep.Error)
}

// Given Id of some Item, return version marked as publish
func (c *natsClient) PublishedVersion(id string) (*wyc.Version, error) {
	req := &codec.PublishedReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return nil, err
	}

	msg, err := c.serverRequest(callGetPublished, data)
	if err != nil {
		return nil, err
	}

	rep := &codec.GetPublishedReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return nil, err
	}

	return &rep.Version, stringError(rep.Error)
}

// Given Id of some Version, mark version as publish
//  - Only one version of a given Item is considered publish at a time
func (c *natsClient) SetPublishedVersion(id string) error {
	req := &codec.PublishedReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.serverRequest(callSetPublished, data)
	if err != nil {
		return err
	}

	rep := &codec.SetPublishedReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return err
	}

	return stringError(rep.Error)
}

// Util func to send update facet request to server & parse reply for errors
func (c *natsClient) genericUpdateFacets(id, subject string, facets map[string]string) error {
	req := &codec.UpdateFacetsReq{
		Id:     id,
		Facets: facets,
	}

	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.serverRequest(subject, data)
	if err != nil {
		return err
	}

	rep := &codec.UpdateFacetsReply{}
	err = rep.UnmarshalJSON(msg.Data)
	if err != nil {
		return err
	}

	return stringError(rep.Error)
}

// Given Collection Id update item facets with given facets
func (c *natsClient) UpdateCollectionFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, callUpdateCollection, facets)
}

// Given Item Id update item facets with given facets
func (c *natsClient) UpdateItemFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, callUpdateItem, facets)
}

// Given Version Id update version facets with given facets
func (c *natsClient) UpdateVersionFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, callUpdateVersion, facets)
}

// Given Resource Id update item facets with given facets
func (c *natsClient) UpdateResourceFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, callUpdateResource, facets)
}

// Given Link Id update item facets with given facets
func (c *natsClient) UpdateLinkFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, callUpdateLink, facets)
}

// wrapper for the server side connection to a nats.io server
type natsServer struct {
	conn     *nats.Conn
	handler  ServerHandler
	subs     []*nats.Subscription
	embedded *natsd.Server
}

// If no nats config is given this is called to spin up a nats server of our own to run embedded.
func (s *natsServer) spinup(options *natsd.Options) (string, error) {
	s.embedded = natsd.New(options)
	go s.embedded.Start()

	if s.embedded.ReadyForConnections(timeout) {
		return fmt.Sprintf("nats://%s:%d", natsDefaultHost, natsDefaultPort), nil
	}
	return "", errors.New("Failed to spin up local nats server")
}

// Start up and serve client requests.
// Incoming client requests will be translated from whatever the middleware protocol is to
// native wysteria objects, then passed to the correct server side handler.
func (s *natsServer) ListenAndServe(config *Settings, handler ServerHandler) error {
	s.handler = handler

	// If we've been told nothing, we'll spin up our own embedded nats server
	if config.Config == "" {
		options := &natsd.Options{
			Host: natsDefaultHost,
			Port: natsDefaultPort,
		}

		if config.SSLEnableTLS {
			options.TLS = config.SSLEnableTLS
			options.TLSVerify = config.SSLVerify
			tlsconf, err := config.TLSconfig()
			if err != nil {
				return err
			}
			options.TLSConfig = tlsconf
		}

		url, err := s.spinup(options)
		if err != nil {
			return err // with no nats to connect to and unable to start one .. we're stuffed
		}
		config.Config = url
	}

	// set up the raw nats.io connection
	raw, err := connect(config)
	if err != nil {
		return err
	}
	s.conn = raw

	// subscribe to all our chans
	fromClients, err := s.subscribe(natsRouteClient+">", natsQueueServer)
	if err != nil {
		return err
	}

	fromServers, err := s.subscribe(natsRouteServer+">", natsQueueServer)
	if err != nil {
		return err
	}

	fromAdmin, err := s.subscribe(natsRouteInternal+">", natsQueueServer)
	if err != nil {
		return err
	}

	// enter the main loop to serve client requests
	for {
		select {
		case message := <-fromClients:
			go s.handleClient(message)
		case message := <-fromServers:
			// ToDo: utilize
			log.Println("server", message)
		case message := <-fromAdmin:
			// ToDo: add admin powers for live management
			log.Println("admin", message)
		}
	}
}

// Util func to obtain the correct length suffix to check against our call<func name> globals
func subjectSuffix(subject string) string {
	return subject[len(subject)-callSuffixLength:]
}

// Util to convert an error to a string so that it can be sent back to the client
func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Assume we've got a create collection request
//  - unmarshal request
//  - call correct server handler func
//  - return whatever the result is
func (s *natsServer) createCollection(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqCollection{}
	err := req.UnmarshalJSON(msg.Data)

	if err == nil {
		// Call handler
		id, err = s.handler.CreateCollection(&req.Collection)
	}
	return &codec.CreateReply{
		Id:    id,
		Error: errorString(err),
	}
}

// Assume we've got a create item request
//  - unmarshal request
//  - call correct server handler func
//  - return whatever the result is
func (s *natsServer) createItem(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqItem{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateItem(&req.Item)
	}
	return &codec.CreateReply{
		Id:    id,
		Error: errorString(err),
	}
}

// Assume we've got a create version request
//  - unmarshal request
//  - call correct server handler func
//  - return whatever the result is
func (s *natsServer) createVersion(msg *nats.Msg) wyc.Marshalable {
	id := ""
	var num int32

	// Unmarshal
	req := codec.CreateReqVersion{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, num, err = s.handler.CreateVersion(&req.Version)
	}
	return &codec.CreateReplyVersion{
		Id:      id,
		Version: num,
		Error:   errorString(err),
	}
}

// Assume we've got a create resource request
//  - unmarshal request
//  - call correct server handler func
//  - return whatever the result is
func (s *natsServer) createResource(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqResource{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateResource(&req.Resource)
	}
	return &codec.CreateReply{
		Id:    id,
		Error: errorString(err),
	}
}

// Assume we've got a create link request
//  - unmarshal request
//  - call correct server handler func
//  - return whatever the result is
func (s *natsServer) createLink(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqLink{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateLink(&req.Link)
	}
	return &codec.CreateReply{
		Id:    id,
		Error: errorString(err),
	}
}

// Generic version of the delete call
//  - unmarshal request (all delete calls supply a single Id)
//  - call correct server handler func
//  - return whatever the result is (all delete calls return an error string)
func (s *natsServer) deleteGeneric(msg *nats.Msg, call func(string) error) wyc.Marshalable {
	// Unmarshal
	req := codec.DeleteReq{}
	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		return &codec.DeleteReply{
			Error: err.Error(),
		}
	}

	// call handler & prepare reply
	return &codec.DeleteReply{
		Error: errorString(call(req.Id)),
	}
}

// Call delete collection
func (s *natsServer) deleteCollection(msg *nats.Msg) wyc.Marshalable {
	return s.deleteGeneric(msg, s.handler.DeleteCollection)
}

// Call delete item
func (s *natsServer) deleteItem(msg *nats.Msg) wyc.Marshalable {
	return s.deleteGeneric(msg, s.handler.DeleteItem)
}

// Call delete version
func (s *natsServer) deleteVersion(msg *nats.Msg) wyc.Marshalable {
	return s.deleteGeneric(msg, s.handler.DeleteVersion)
}

// Call delete resource
func (s *natsServer) deleteResource(msg *nats.Msg) wyc.Marshalable {
	return s.deleteGeneric(msg, s.handler.DeleteResource)
}

// Assume we've got a find collection request
//  - unmarshal request
//  - call server handler func
//  - marshal & return result to client
func (s *natsServer) findCollection(msg *nats.Msg) wyc.Marshalable {
	req := &codec.FindReq{}
	rep := &codec.FindReplyCollection{
		All: []wyc.Collection{},
	}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	qs := []*wyc.QueryDesc{}
	for _, q := range req.Query {
		tmp := q
		qs = append(qs, &tmp)
	}

	result, err := s.handler.FindCollections(req.Limit, req.Offset, qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

// Assume we've got a find item request
//  - unmarshal request
//  - call server handler func
//  - marshal & return result to client
func (s *natsServer) findItem(msg *nats.Msg) wyc.Marshalable {
	req := &codec.FindReq{}
	rep := &codec.FindReplyItem{
		All: []wyc.Item{},
	}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	qs := []*wyc.QueryDesc{}
	for _, q := range req.Query {
		tmp := q
		qs = append(qs, &tmp)
	}

	result, err := s.handler.FindItems(req.Limit, req.Offset, qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

// Assume we've got a find version request
//  - unmarshal request
//  - call server handler func
//  - marshal & return result to client
func (s *natsServer) findVersion(msg *nats.Msg) wyc.Marshalable {
	req := &codec.FindReq{}
	rep := &codec.FindReplyVersion{
		All: []wyc.Version{},
	}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	qs := []*wyc.QueryDesc{}
	for _, q := range req.Query {
		tmp := q
		qs = append(qs, &tmp)
	}

	result, err := s.handler.FindVersions(req.Limit, req.Offset, qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

// Assume we've got a find resource request
//  - unmarshal request
//  - call server handler func
//  - marshal & return result to client
func (s *natsServer) findResource(msg *nats.Msg) wyc.Marshalable {
	req := &codec.FindReq{}
	rep := &codec.FindReplyResource{
		All: []wyc.Resource{},
	}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	qs := []*wyc.QueryDesc{}
	for _, q := range req.Query {
		tmp := q
		qs = append(qs, &tmp)
	}

	result, err := s.handler.FindResources(req.Limit, req.Offset, qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

// Assume we've got a find link request
//  - unmarshal request
//  - call server handler func
//  - marshal & return result to client
func (s *natsServer) findLink(msg *nats.Msg) wyc.Marshalable {
	req := &codec.FindReq{}
	rep := &codec.FindReplyLink{
		All: []wyc.Link{},
	}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	qs := []*wyc.QueryDesc{}
	for _, q := range req.Query {
		tmp := q
		qs = append(qs, &tmp)
	}

	result, err := s.handler.FindLinks(req.Limit, req.Offset, qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

// Assume we've got a request to get a publish version
//  - unmarshal request
//  - call server handler func
//  - marshal & return result
func (s *natsServer) publishedVersion(msg *nats.Msg) wyc.Marshalable {
	req := &codec.PublishedReq{}
	rep := &codec.GetPublishedReply{}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	version, err := s.handler.PublishedVersion(req.Id)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	rep.Version = *version

	return rep
}

// Assume we've got a request to set a version as publish
//  - unmarshal request
//  - call server handler func
//  - marshal & return result
func (s *natsServer) setPublished(msg *nats.Msg) wyc.Marshalable {
	req := &codec.PublishedReq{}
	rep := &codec.CreateReplyVersion{}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	rep.Error = errorString(s.handler.SetPublishedVersion(req.Id))
	return rep
}

// Util func to update the facets of an item / version
//  - unmarshal request
//  - call server handler func
//  - marshal & return result
func (s *natsServer) genericUpdateFacets(msg *nats.Msg, call func(string, map[string]string) error) wyc.Marshalable {
	req := &codec.UpdateFacetsReq{}
	rep := &codec.UpdateFacetsReply{}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	rep.Error = errorString(call(
		req.Id,
		req.Facets,
	))
	return rep
}

// Update collection facets from client request msg
func (s *natsServer) updateCollection(msg *nats.Msg) wyc.Marshalable {
	return s.genericUpdateFacets(msg, s.handler.UpdateCollectionFacets)
}

// Update version facets from client request msg
func (s *natsServer) updateVersion(msg *nats.Msg) wyc.Marshalable {
	return s.genericUpdateFacets(msg, s.handler.UpdateVersionFacets)
}

// Update item facets from client request msg
func (s *natsServer) updateItem(msg *nats.Msg) wyc.Marshalable {
	return s.genericUpdateFacets(msg, s.handler.UpdateItemFacets)
}

// Update resource facets from client request msg
func (s *natsServer) updateResource(msg *nats.Msg) wyc.Marshalable {
	return s.genericUpdateFacets(msg, s.handler.UpdateResourceFacets)
}

// Update link facets from client request msg
func (s *natsServer) updateLink(msg *nats.Msg) wyc.Marshalable {
	return s.genericUpdateFacets(msg, s.handler.UpdateLinkFacets)
}

// Send reply to client ..
// Hopefully we can marshal our answer into what we're expecting, but we'll
// try really hard to send SOMETHING even if it isn't what we hoped to. The
// idea is that the client (even on an err) will have something helpful to
// print / understand what happened if nothing else.
func (s *natsServer) sendReply(to string, m wyc.Marshalable) {
	// Marshal reply
	data, err := m.MarshalJSON()
	if err != nil {
		log.Println("error in sendReply [MarshalJSON]", err, "given", m)
		s.natsPublish(to, []byte(err.Error()))
		return
	}

	// Send reply
	err = s.natsPublish(to, data)
	if err != nil {
		log.Println("error in sendReply [publish]", err, "given", m)
		s.natsPublish(to, []byte(err.Error()))
		return
	}
}

// Choose the correct func to unmarshal/marshal client request given the subject of a message we've received.
// That is, we're going to try and match the subject to one of our pre-configured strings.
func (s *natsServer) chooseClientHandler(subject string) func(*nats.Msg) wyc.Marshalable {
	var handler func(*nats.Msg) wyc.Marshalable

	switch subjectSuffix(subject) {
	case callCreateCollection:
		handler = s.createCollection
	case callCreateItem:
		handler = s.createItem
	case callCreateVersion:
		handler = s.createVersion
	case callCreateResource:
		handler = s.createResource
	case callCreateLink:
		handler = s.createLink

	case callDeleteCollection:
		handler = s.deleteCollection
	case callDeleteItem:
		handler = s.deleteItem
	case callDeleteVersion:
		handler = s.deleteVersion
	case callDeleteResource:
		handler = s.deleteResource

	case callFindCollection:
		handler = s.findCollection
	case callFindItem:
		handler = s.findItem
	case callFindVersion:
		handler = s.findVersion
	case callFindResource:
		handler = s.findResource
	case callFindLink:
		handler = s.findLink

	case callGetPublished:
		handler = s.publishedVersion
	case callSetPublished:
		handler = s.setPublished

	case callUpdateCollection:
		handler = s.updateCollection
	case callUpdateItem:
		handler = s.updateItem
	case callUpdateVersion:
		handler = s.updateVersion
	case callUpdateResource:
		handler = s.updateResource
	case callUpdateLink:
		handler = s.updateLink
	}

	return handler
}

// Pick out the correct function to call given the received request.
// This func will handle parsing the request, calling the handler and returning the correct reply format
// the client will be expecting.
//
// If we don't find a match or the unmarshal fails, we'll simply fire the error back to the client.
// This will probably cause a failure client side, but if we've no matching func it really means the client
// is not written correctly and needs to be fixed.
func (s *natsServer) handleClient(msg *nats.Msg) {
	handler := s.chooseClientHandler(msg.Subject)

	if handler == nil {
		log.Println("Handler not found for", msg.Subject)
		return
	}

	// Here we call the handler - it can do whatever and return us
	// some kind of wyc.Marshalable result
	result := handler(msg)

	// Pass whatever the result was to our func to be turned back into
	// a []byte and sent off.
	s.sendReply(msg.Reply, result)
}

// Connect to nats a server given the settings object
//  Url required by nats.io (from the docs)
//    nats://derek:pass@localhost:4222
//  As in,
//    nats://user:password@host:port
// This also wraps up configuring the nats connection with TLS & similar options
//
func connect(config *Settings) (*nats.Conn, error) {
	opts := []nats.Option{}
	if config.SSLEnableTLS {
		tlsconf, err := config.TLSconfig()
		if err != nil {
			return nil, err
		}
		opts = append(opts, nats.Secure(tlsconf))
	}

	return nats.Connect(config.Config, opts...)
}

// Shutdown the server and kill connections
func (s *natsServer) Shutdown() error {
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}

	s.conn.Flush()
	s.conn.Close()
	return nil
}

// Create a new middleware server endpoint
func newNatsServer() EndpointServer {
	return &natsServer{
		subs: []*nats.Subscription{},
	}
}

// Publish
//   Send message over nats and don't wait for a reply
func (s *natsServer) natsPublish(subject string, data []byte) error {
	return s.conn.Publish(subject, data)
}

// Subscribe to channel
//   A message sent to the chan is received by only one listener
func (s *natsServer) subscribe(subject, queue string) (<-chan *nats.Msg, error) {
	recv := make(chan *nats.Msg) // make an inbound chan
	f := func(msg *nats.Msg) {
		recv <- msg
	}

	sub, err := s.conn.QueueSubscribe(subject, queue, f)
	if err != nil {
		return nil, err
	}
	s.subs = append(s.subs, sub)

	return recv, nil
}
