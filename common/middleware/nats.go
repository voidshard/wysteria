package middleware

import (
	wyc "github.com/voidshard/wysteria/common"
	"github.com/nats-io/nats"
	natsd "github.com/nats-io/gnatsd/server"
	"time"
	"log"
	codec "github.com/voidshard/wysteria/common/middleware/ffjson_structs"
	"errors"
	"fmt"
)

const (
	nats_default_host = "localhost"
	nats_default_port = 4222
	nats_queue_server = "server_queue"

	// These routes indicate who is sending the message
	nats_route_server = "w.server."  // From a wysteria server
	nats_route_client = "w.client."   // From a client
	nats_route_internal = "w.internal."  // From the admin(s)

	// messages suffixes
	call_suffix_length = 2
	call_create_collection = "cc"
	call_create_item = "ci"
	call_create_version = "cv"
	call_create_resource = "cr"
	call_create_link = "cl"

	call_delete_collection = "dc"
	call_delete_item = "di"
	call_delete_version = "dv"
	call_delete_resource = "dr"

	call_find_collection = "fc"
	call_find_item = "fi"
	call_find_version = "fv"
	call_find_resource = "fr"
	call_find_link = "fl"

	call_get_published = "gp"
	call_set_published = "sp"

	call_update_version = "uv"
	call_update_item = "ui"
)

var (
	timeout = time.Second * 30
)

type natsClient struct {
	conn    *nats.Conn
}

func newNatsClient() EndpointClient {
	return &natsClient{}
}

// Connect to nats given the url
//  Url required by nats.io (from the docs)
//    nats://derek:pass@localhost:4222
//  As in,
//    nats://user:password@host:port
func (c *natsClient) Connect(config string) error {
	if config == "" {
		config = fmt.Sprintf("nats://%s:%d", nats_default_host, nats_default_port)
	}

	raw, err := nats.Connect(config)
	if err != nil {
		return err
	}
	c.conn = raw
	return err
}

func (c *natsClient) server_request(subject string, data []byte) (*nats.Msg, error) {
	return c.conn.Request(nats_route_client + subject, data, timeout)
}

func (c *natsClient) Close() error {
	c.conn.Flush()
	c.conn.Close()
	return nil
}

func stringError(in string) error {
	if in == "" {
		return nil
	}
	return errors.New(in)
}

func (c *natsClient) CreateCollection(name string) (id string, err error) {
	req := &codec.CreateReqCollection{Name: name}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_create_collection, data)
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

func (c *natsClient) CreateItem(in *wyc.Item) (id string, err error) {
	req := &codec.CreateReqItem{Item: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_create_item, data)
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

func (c *natsClient) CreateVersion(in *wyc.Version) (id string, num int32, err error){
	req := &codec.CreateReqVersion{Version: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_create_version, data)
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

func (c *natsClient) CreateResource(in *wyc.Resource) (id string, err error) {
	req := &codec.CreateReqResource{Resource: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_create_resource, data)
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

func (c *natsClient) CreateLink(in *wyc.Link) (id string, err error) {
	req := &codec.CreateReqLink{Link: *in}
	data, err := req.MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_create_link, data)
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

func (c *natsClient) genericDelete(id, subject string) error {
	req := &codec.DeleteReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.server_request(subject, data)
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

func (c *natsClient) DeleteCollection(id string) error {
	return c.genericDelete(id, call_delete_collection)
}

func (c *natsClient) DeleteItem(id string) error {
	return c.genericDelete(id, call_delete_item)
}

func (c *natsClient) DeleteVersion(id string) error {
	return c.genericDelete(id, call_delete_version)
}

func (c *natsClient) DeleteResource(id string) error {
	return c.genericDelete(id, call_delete_resource)
}

func toFindReq(query []*wyc.QueryDesc) (*codec.FindReq) {
	req := &codec.FindReq{
		Query: []wyc.QueryDesc{},
	}
	for _, q := range query {
		req.Query = append(req.Query, *q)
	}
	return req
}

func (c *natsClient) FindCollections(query []*wyc.QueryDesc) (result []*wyc.Collection, err error) {
	data, err := toFindReq(query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_find_collection, data)
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

func (c *natsClient) FindItems(query []*wyc.QueryDesc) (result []*wyc.Item, err error) {
	data, err := toFindReq(query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_find_item, data)
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
func (c *natsClient) FindVersions(query []*wyc.QueryDesc) (result []*wyc.Version, err error) {
	data, err := toFindReq(query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_find_version, data)
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
func (c *natsClient) FindResources(query []*wyc.QueryDesc) (result []*wyc.Resource, err error) {
	data, err := toFindReq(query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_find_resource, data)
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

func (c *natsClient) FindLinks(query []*wyc.QueryDesc) (result []*wyc.Link, err error) {
	data, err := toFindReq(query).MarshalJSON()
	if err != nil {
		return
	}

	msg, err := c.server_request(call_find_link, data)
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

func (c *natsClient) GetPublishedVersion(id string) (*wyc.Version, error) {
	req := &codec.PublishedReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return nil, err
	}

	msg, err := c.server_request(call_get_published, data)
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

func (c *natsClient) PublishVersion(id string) error {
	req := &codec.PublishedReq{Id: id}
	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.server_request(call_set_published, data)
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

func (c *natsClient) genericUpdateFacets(id, subject string, facets map[string]string) error {
	req := &codec.UpdateFacetsReq{
		Id: id,
		Facets: facets,
	}

	data, err := req.MarshalJSON()
	if err != nil {
		return err
	}

	msg, err := c.server_request(subject, data)
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

func (c *natsClient) UpdateItemFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, call_update_item, facets)
}

func (c *natsClient) UpdateVersionFacets(id string, facets map[string]string) error {
	return c.genericUpdateFacets(id, call_update_version, facets)
}

type natsServer struct {
	conn   *nats.Conn
	handler ServerHandler
	subs []*nats.Subscription
	embedded *natsd.Server
}

func (s *natsServer) spinup() (string, error) {
	s.embedded = natsd.New(&natsd.Options{
		Host: nats_default_host,
		Port: nats_default_port,
	})
	go s.embedded.Start()

	if s.embedded.ReadyForConnections(timeout) {
		return fmt.Sprintf("nats://%s:%d", nats_default_host, nats_default_port), nil
	}
	return "", errors.New("Failed to spin up local nats server")
}

// Start up and serve client requests
func (s *natsServer) ListenAndServe(config string, handler ServerHandler) error {
	s.handler = handler

	// If we've been told nothing, spin up our own nats
	if config == "" {
		url, err := s.spinup()
		if err != nil {
			return err
		}
		config = url
	}

	// set up the raw nats.io connection
	err := s.connect(config)
	if err != nil {
		return err
	}

	// subscribe to all our chans
	fromClients, err := s.subscribe(nats_route_client + ">", nats_queue_server)
	if err != nil {
		return err
	}

	fromServers, err := s.subscribe(nats_route_server + ">", nats_queue_server)
	if err != nil {
		return err
	}

	fromAdmin, err := s.subscribe(nats_route_internal + ">", nats_queue_server)
	if err != nil {
		return err
	}

	// enter the loop
	for {
		select {
		case message := <- fromClients:
			go s.handle_client(message)
		case message := <- fromServers:
			// ToDo: utilize
			log.Println("server", message)
		case message := <- fromAdmin:
			// ToDo: add admin powers for live management
			log.Println("admin", message)
		}
	}
}

func subjectSuffix(subject string) string {
	return subject[len(subject) - call_suffix_length:]
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func (s *natsServer) create_collection(msg *nats.Msg) wyc.Marshalable {
	id := ""
	
	// Unmarshal
	req := codec.CreateReqCollection{}
	err := req.UnmarshalJSON(msg.Data)

	if err == nil {
		// Call handler
		id, err = s.handler.CreateCollection(req.Name)
	}
	return &codec.CreateReply{
		Id: id,
		Error: errorString(err),
	}
}

func (s *natsServer) create_item(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqItem{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateItem(&req.Item)
	}
	return &codec.CreateReply{
		Id: id,
		Error: errorString(err),
	}
}

func (s *natsServer) create_version(msg *nats.Msg) wyc.Marshalable {
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
		Id: id,
		Version: num,
		Error: errorString(err),
	}
}

func (s *natsServer) create_resource(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqResource{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateResource(&req.Resource)
	}
	return &codec.CreateReply{
		Id: id,
		Error: errorString(err),
	}
}

func (s *natsServer) create_link(msg *nats.Msg) wyc.Marshalable {
	id := ""

	// Unmarshal
	req := codec.CreateReqLink{}
	err := req.UnmarshalJSON(msg.Data)

	// Call handler
	if err == nil {
		id, err = s.handler.CreateLink(&req.Link)
	}
	return &codec.CreateReply{
		Id: id,
		Error: errorString(err),
	}
}

// Generic version of the delete call
func (s *natsServer) delete_generic(msg *nats.Msg, call func(string) error) wyc.Marshalable {
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

func (s *natsServer) delete_collection(msg *nats.Msg) wyc.Marshalable {
	return s.delete_generic(msg, s.handler.DeleteCollection)
}

func (s *natsServer) delete_item(msg *nats.Msg) wyc.Marshalable {
	return s.delete_generic(msg, s.handler.DeleteItem)
}

func (s *natsServer) delete_version(msg *nats.Msg) wyc.Marshalable {
	return s.delete_generic(msg, s.handler.DeleteVersion)
}

func (s *natsServer) delete_resource(msg *nats.Msg) wyc.Marshalable {
	return s.delete_generic(msg, s.handler.DeleteResource)
}

func (s *natsServer) find_collection(msg *nats.Msg) wyc.Marshalable {
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

	result, err := s.handler.FindCollections(qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

func (s *natsServer) find_item(msg *nats.Msg) wyc.Marshalable {
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

	result, err := s.handler.FindItems(qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

func (s *natsServer) find_version(msg *nats.Msg) wyc.Marshalable {
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

	result, err := s.handler.FindVersions(qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

func (s *natsServer) find_resource(msg *nats.Msg) wyc.Marshalable {
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

	result, err := s.handler.FindResources(qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

func (s *natsServer) find_link(msg *nats.Msg) wyc.Marshalable {
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

	result, err := s.handler.FindLinks(qs)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	for _, r := range result {
		rep.All = append(rep.All, *r)
	}
	return rep
}

func (s *natsServer) get_published(msg *nats.Msg) wyc.Marshalable {
	req := &codec.PublishedReq{}
	rep := &codec.GetPublishedReply{}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	version, err := s.handler.GetPublishedVersion(req.Id)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}
	rep.Version = *version

	return rep
}

func (s *natsServer) set_published(msg *nats.Msg) wyc.Marshalable {
	req := &codec.PublishedReq{}
	rep := &codec.CreateReplyVersion{}

	err := req.UnmarshalJSON(msg.Data)
	if err != nil {
		rep.Error = err.Error()
		return rep
	}

	rep.Error = errorString(s.handler.PublishVersion(req.Id))
	return rep
}

func (s *natsServer) generic_update_facets(msg *nats.Msg, call func(string, map[string]string) error) wyc.Marshalable {
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

func (s *natsServer) update_version(msg *nats.Msg) wyc.Marshalable {
	return s.generic_update_facets(msg, s.handler.UpdateVersionFacets)
}

func (s *natsServer) update_item(msg *nats.Msg) wyc.Marshalable {
	return s.generic_update_facets(msg, s.handler.UpdateItemFacets)
}

// Send reply to client ..
// Hopefully we can marshal our answer into what we're expecting, but we'll
// try really hard to send SOMETHING even if it isn't what we hoped to. The
// idea is that the client (even on an err) will have something helpful to
// print / understand what happened if nothing else.
//
func (s *natsServer) send_reply(to string, m wyc.Marshalable) {
	// Marshal reply
	data, err := m.MarshalJSON()
	if err != nil {
		log.Println("error in send_reply [MarshalJSON]", err, "given", m)
		s.publish(to, []byte(err.Error()))
		return
	}

	// Send reply
	err = s.publish(to, data)
	if err != nil {
		log.Println("error in send_reply [publish]", err, "given", m)
		s.publish(to, []byte(err.Error()))
		return
	}
}

func (s *natsServer) choose_client_handler(subject string) func(*nats.Msg) wyc.Marshalable {
	var handler func(*nats.Msg) wyc.Marshalable

	switch subjectSuffix(subject) {
	case call_create_collection:
		handler = s.create_collection
	case call_create_item:
		handler = s.create_item
	case call_create_version:
		handler = s.create_version
	case call_create_resource:
		handler = s.create_resource
	case call_create_link:
		handler = s.create_link

	case call_delete_collection:
		handler = s.delete_collection
	case call_delete_item:
		handler = s.delete_item
	case call_delete_version:
		handler = s.delete_version
	case call_delete_resource:
		handler = s.delete_resource

	case call_find_collection:
		handler = s.find_collection
	case call_find_item:
		handler = s.find_item
	case call_find_version:
		handler = s.find_version
	case call_find_resource:
		handler = s.find_resource
	case call_find_link:
		handler = s.find_link

	case call_get_published:
		handler = s.get_published
	case call_set_published:
		handler = s.set_published
	case call_update_item:
		handler = s.update_item
	case call_update_version:
		handler = s.update_version
	}

	return handler
}

func (s *natsServer) handle_client(msg *nats.Msg) {
	// Pick out the correct function to call. This handles parsing
	// the req, calling the handler and returning the correct reply format
	handler := s.choose_client_handler(msg.Subject)

	if handler == nil {
		log.Println("Handler not found for", msg.Subject)
		return
	}

	// Here we call the handler - it can do whatever and return us 
	// some kind of wyc.Marshalable result
	result := handler(msg)
	
	// Pass whatever the result was to our func to be turned back into
	// a []byte and sent off. 
	s.send_reply(msg.Reply, result)
}

// Connect to nats given the url
//  Url required by nats.io (from the docs)
//    nats://derek:pass@localhost:4222
//  As in,
//    nats://user:password@host:port
//
func (s *natsServer) connect(config string) error {
	raw, err := nats.Connect(config)
	if err != nil {
		return err
	}
	s.conn = raw
	return err
}

func (s *natsServer) Shutdown() error {
	for _, sub := range s.subs {
		sub.Unsubscribe()
	}

	s.conn.Flush()
	s.conn.Close()
	return nil
}

func newNatsServer() EndpointServer {
	return &natsServer{
		subs: []*nats.Subscription{},
	}
}

// Publish
//   Send message and don't wait for a reply
//
func (s *natsServer) publish(subject string, data []byte) error {
	return s.conn.Publish(subject, data)
}

// Subscribe to channel
//   A message sent to the chan is received by only one listener
//
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
