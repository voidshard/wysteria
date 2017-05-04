/*
The gRPC protocol here is implemented mostly for local tests and as an example of implementing custom middleware(s)
into Wysteria. It's my suspicion that Nats.io is better in every way :P

Still, this does have the benefit of protobuf writing most of the code for us ..
*/

package middleware

import (
	"errors"
	wyc "github.com/voidshard/wysteria/common"
	wrpc "github.com/voidshard/wysteria/common/middleware/grpc_proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"net"
)

var (
	// These are 'Null' answers used when we're send an error message and shouldn't be written to ..
	nullWrpcId       = &wrpc.Id{Id: ""}
	nullWrpcIdAndNum = &wrpc.IdAndNum{Id: "", Version: 0}
	nullWrpcText     = &wrpc.Text{Text: ""}
)

// Our wrapper client around the auto generated protobuf client to provide nicer interaction
type grpcClient struct {
	config string
	conn   *grpc.ClientConn
	client wrpc.WysteriaGrpcClient
}

// Create a new client & connect to the remote RPC server
func (c *grpcClient) Connect(config string) error {
	if config == "" {
		config = "localhost:50051"
	}
	conn, err := grpc.Dial(config, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.conn = conn

	c.client = wrpc.NewWysteriaGrpcClient(c.conn)
	return nil
}

// Close connection(s) to the server
func (c *grpcClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// Call the server side CreateCollection func with the given 'name'
// That is, create a new Collection with the name given
func (c *grpcClient) CreateCollection(name string) (string, error) {
	result, err := c.client.CreateCollection(
		context.Background(),
		&wrpc.Text{Text: name},
	)
	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

// util func to convert a wysteria.Item (native) into a wrpc.Item (middleware)
func convWItem(in *wyc.Item) *wrpc.Item {
	return &wrpc.Item{
		Id:       in.Id,
		Parent:   in.Parent,
		ItemType: in.ItemType,
		Variant:  in.Variant,
		Facets:   in.Facets,
	}
}

// Create a new Item, based on the given Item
func (c *grpcClient) CreateItem(in *wyc.Item) (string, error) {
	result, err := c.client.CreateItem(
		context.Background(),
		convWItem(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

// Util func to convert wysteria native Version obj to an rpc Version
func convWVersion(in *wyc.Version) *wrpc.Version {
	return &wrpc.Version{
		Id:     in.Id,
		Parent: in.Parent,
		Number: in.Number,
		Facets: in.Facets,
	}
}

// Create a new Version based on the given input
func (c *grpcClient) CreateVersion(in *wyc.Version) (string, int32, error) {
	result, err := c.client.CreateVersion(
		context.Background(),
		convWVersion(in),
	)

	if err != nil {
		return "", 0, err
	}
	if result.Text != "" {
		return "", 0, errors.New(result.Text)
	}
	return result.Id, result.Version, nil
}

// Util func to convert wysteria native Resource obj into an rpc Resource obj
func convWResource(in *wyc.Resource) *wrpc.Resource {
	return &wrpc.Resource{
		Id:           in.Id,
		Parent:       in.Parent,
		Name:         in.Name,
		Location:     in.Location,
		ResourceType: in.ResourceType,
	}
}

// Create a new Resource based on the given input
func (c *grpcClient) CreateResource(in *wyc.Resource) (string, error) {
	result, err := c.client.CreateResource(
		context.Background(),
		convWResource(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

// Util to convert wysteria native Link obj to rpc Link obj
func convWLink(in *wyc.Link) *wrpc.Link {
	return &wrpc.Link{
		Id:   in.Id,
		Src:  in.Src,
		Name: in.Name,
		Dst:  in.Dst,
	}
}

// Create a new Link based on the given input
func (c *grpcClient) CreateLink(in *wyc.Link) (string, error) {
	result, err := c.client.CreateLink(
		context.Background(),
		convWLink(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error != nil {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

// Util func that calls some server side delete function, passing in the Id and returning the error
func clientCallDelete(delete_id string, call func(ctx context.Context, in *wrpc.Id, opts ...grpc.CallOption) (*wrpc.Text, error)) error {
	result, err := call(context.Background(), &wrpc.Id{Id: delete_id})
	if err != nil {
		return err
	}
	if result.Text != "" {
		return errors.New(result.Text)
	}
	return nil
}

// Delete a Collection, given it's Id
func (c *grpcClient) DeleteCollection(in string) error {
	return clientCallDelete(in, c.client.DeleteCollection)
}

// Delete a Item, given it's Id
func (c *grpcClient) DeleteItem(in string) error {
	return clientCallDelete(in, c.client.DeleteItem)
}

// Delete a Version, given it's Id
func (c *grpcClient) DeleteVersion(in string) error {
	return clientCallDelete(in, c.client.DeleteVersion)
}

// Delete a Resource, given it's Id
func (c *grpcClient) DeleteResource(in string) error {
	return clientCallDelete(in, c.client.DeleteResource)
}

// Util func to convert wysteria native QueryDesc objects to rpc QueryDesc objects
func convWQueryDescs(in ...*wyc.QueryDesc) *wrpc.QueryDescs {
	result := []*wrpc.QueryDesc{}
	for _, q := range in {
		result = append(
			result,
			&wrpc.QueryDesc{
				Parent:        q.Parent,
				Id:            q.Id,
				VersionNumber: q.VersionNumber,
				ItemType:      q.ItemType,
				Variant:       q.Variant,
				Facets:        q.Facets,
				Name:          q.Name,
				ResourceType:  q.ResourceType,
				Location:      q.Location,
				LinkSrc:       q.LinkSrc,
				LinkDst:       q.LinkDst,
			},
		)
	}
	return &wrpc.QueryDescs{All: result}
}

// Util func to convert rpc Collection objs to native wysteria Collection objects
func convRCollections(in ...*wrpc.Collection) []*wyc.Collection {
	result := []*wyc.Collection{}
	for _, i := range in {
		result = append(result, &wyc.Collection{
			Id:   i.Id,
			Name: i.Name,
		})
	}
	return result
}

// Given the input queries, find & return all matching Collections
func (c *grpcClient) FindCollections(in []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	result, err := c.client.FindCollections(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRCollections(result.All...), nil
}

// Util func to convert rpc Item objs to native wysteria Item objects
func convRItems(in ...*wrpc.Item) []*wyc.Item {
	result := []*wyc.Item{}
	for _, i := range in {
		result = append(result, &wyc.Item{
			Id:       i.Id,
			Parent:   i.Parent,
			ItemType: i.ItemType,
			Variant:  i.Variant,
			Facets:   i.Facets,
		})
	}
	return result
}

// Given the input queries, find & return all matching Items
func (c *grpcClient) FindItems(in []*wyc.QueryDesc) ([]*wyc.Item, error) {
	result, err := c.client.FindItems(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRItems(result.All...), nil
}

// Util func to convert rpc Version objs to native wysteria Version objects
func convRVersions(in ...*wrpc.Version) []*wyc.Version {
	result := []*wyc.Version{}
	for _, i := range in {
		result = append(result, &wyc.Version{
			Id:     i.Id,
			Parent: i.Parent,
			Facets: i.Facets,
			Number: i.Number,
		})
	}
	return result
}

// Given the input queries, find & return all matching Versions
func (c *grpcClient) FindVersions(in []*wyc.QueryDesc) ([]*wyc.Version, error) {
	result, err := c.client.FindVersions(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRVersions(result.All...), nil
}

// Util func to convert rpc Resource objs to native wysteria Resource objects
func convRResources(in ...*wrpc.Resource) []*wyc.Resource {
	result := []*wyc.Resource{}
	for _, i := range in {
		result = append(result, &wyc.Resource{
			Id:           i.Id,
			Parent:       i.Parent,
			Name:         i.Name,
			ResourceType: i.ResourceType,
			Location:     i.Location,
		})
	}
	return result
}

// Given the input queries, find & return all matching Resource
func (c *grpcClient) FindResources(in []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	result, err := c.client.FindResources(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRResources(result.All...), nil
}

// Util func to convert rpc Link objs to native wysteria Link objects
func convRLinks(in ...*wrpc.Link) []*wyc.Link {
	result := []*wyc.Link{}
	for _, i := range in {
		result = append(result, &wyc.Link{
			Id:   i.Id,
			Name: i.Name,
			Src:  i.Src,
			Dst:  i.Dst,
		})
	}
	return result
}

// Given the input queries, find & return all matching Link
func (c *grpcClient) FindLinks(in []*wyc.QueryDesc) ([]*wyc.Link, error) {
	result, err := c.client.FindLinks(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRLinks(result.All...), nil
}

// Util func to convert rpc Version objs to native wysteria Version objects
func convRVersion(in *wrpc.Version) *wyc.Version {
	return &wyc.Version{
		Id:     in.Id,
		Parent: in.Parent,
		Facets: in.Facets,
		Number: in.Number,
	}
}

// Given the Id of the parent Item, return the published Version
func (c *grpcClient) GetPublishedVersion(in string) (*wyc.Version, error) {
	result, err := c.client.GetPublishedVersion(context.Background(), &wrpc.Id{Id: in})
	if err != nil {
		return nil, err
	}

	if result.Error != nil {
		return nil, errors.New(result.Error.Text)
	}
	return convRVersion(result), nil
}

// Mark the Version with the given Id as the current 'published' one
func (c *grpcClient) PublishVersion(in string) error {
	result, err := c.client.PublishVersion(context.Background(), &wrpc.Id{Id: in})
	if err != nil {
		return err
	}

	if result.Text != "" {
		return errors.New(result.Text)
	}
	return nil
}

// Util func to call some update facets function and return the error
func callUpdate(id string, facets map[string]string, call func(context.Context, *wrpc.IdAndDict, ...grpc.CallOption) (*wrpc.Text, error)) error {
	result, err := call(context.Background(), &wrpc.IdAndDict{Id: id, Facets: facets})
	if err != nil {
		return err
	}
	if result.Text != "" {
		return errors.New(result.Text)
	}
	return nil
}

// Update the facets of the Version with the given Id with the given facets
func (c *grpcClient) UpdateVersionFacets(id string, to_update map[string]string) error {
	return callUpdate(id, to_update, c.client.UpdateVersionFacets)
}

// Update the facets of the Item with the given Id with the given facets
func (c *grpcClient) UpdateItemFacets(id string, to_update map[string]string) error {
	return callUpdate(id, to_update, c.client.UpdateItemFacets)
}

// Create a new gRPC client and return it
func newGrpcClient() EndpointClient {
	return &grpcClient{}
}

// Wrapper around the raw machine generated protobuf rpc server
type grpcServer struct {
	conn    net.Listener
	config  string
	server  wrpc.WysteriaGrpcServer
	handler ServerHandler
}

// Create a new gRPC server
func newGrpcServer() EndpointServer {
	return &grpcServer{}
}

// ListenAndServe grpc requests
//  Essentially what we're doing here is wiring together three interfaces
//
//   Server (grpc/server.go):
//    the Go RPC server which receives client requests
//
//   ServerHandler (wysteria/common/middleware/middleware.go)
//    the internal wysteria server interface that requests from the middleware will be routed through
//    to actually reach the server layer and do the given work
//
//   grpcServer (wysteria/common/middleware/grpc.go)
//    implements the wysteria/common/middleware/grpc_proto/wysteria.grpc.pb.go grpc service
//    defined in the .proto file. This is what receives requests from the google grpc package "Server"
//    and changes data from the grpc message format(s) to our own objects (where required).
//    This obj back and forth is kinda inefficient but allows us to implement nice interfaces everywhere.
//
//   That is, incoming requests go
//  Client -> Server (grpc server) -> grpcServer (our middleware interface) -> ServerHandler (main server interface)
//   Then back out
//  ServerHandler -> grpcServer -> Server -> Client
//   (Where the client is running the rpc client wrapped by our grpcClient implementation)
//
//  Note that all of the functions in grpcServer essentially turn rpc message objects into wysteria common
//  objects, pass them into the correct ServerHandler function(s) then return rpc message objects again.
//
// It's simple really, if you don't think about it.
//
func (s *grpcServer) ListenAndServe(config string, handler ServerHandler) error {
	if config == "" {
		config = ":50051"
	}

	conn, err := net.Listen("tcp", config)
	if err != nil {
		return err
	}

	s.conn = conn
	s.handler = handler

	rpc_server := grpc.NewServer()
	wrpc.RegisterWysteriaGrpcServer(rpc_server, s)

	// Register reflection service on gRPC server.
	reflection.Register(rpc_server)

	return rpc_server.Serve(s.conn)
}

// Close any and all connections
func (s *grpcServer) Shutdown() error {
	if s.server == nil {
		return nil
	}
	return s.conn.Close()
}

// Given the Id and Facets to update, call the server side UpdateVersionFacets func, return result
func (s *grpcServer) UpdateVersionFacets(_ context.Context, in *wrpc.IdAndDict) (*wrpc.Text, error) {
	err := s.handler.UpdateVersionFacets(in.Id, in.Facets)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return nullWrpcText, nil
}

// Given the Id and Facets to update, call the server side UpdateItemFacets func, return result
func (s *grpcServer) UpdateItemFacets(_ context.Context, in *wrpc.IdAndDict) (*wrpc.Text, error) {
	err := s.handler.UpdateItemFacets(in.Id, in.Facets)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return nullWrpcText, nil
}

// Given the name of the collection to create, call server side CreateCollection, return result
func (s *grpcServer) CreateCollection(_ context.Context, in *wrpc.Text) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateCollection(in.Text)
	if err != nil {
		return nullWrpcId, err
	}
	return &wrpc.Id{Id: created_id}, err
}

// Given base obj, call server side CreateItem, return result
func (s *grpcServer) CreateItem(_ context.Context, in *wrpc.Item) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateItem(convRItems(in)[0])
	if err != nil {
		return nullWrpcId, err
	}
	return &wrpc.Id{Id: created_id}, err
}

// Given base obj, call server side CreateVersion, return result
func (s *grpcServer) CreateVersion(_ context.Context, in *wrpc.Version) (*wrpc.IdAndNum, error) {
	created_id, number, err := s.handler.CreateVersion(convRVersion(in))
	if err != nil {
		return nullWrpcIdAndNum, err
	}
	return &wrpc.IdAndNum{Id: created_id, Version: number}, err
}

// Given base obj, call server side CreateResource, return result
func (s *grpcServer) CreateResource(_ context.Context, in *wrpc.Resource) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateResource(convRResources(in)[0])
	if err != nil {
		return nullWrpcId, err
	}
	return &wrpc.Id{Id: created_id}, err
}

// Given base obj, call server side CreateLink, return result
func (s *grpcServer) CreateLink(_ context.Context, in *wrpc.Link) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateLink(convRLinks(in)[0])
	if err != nil {
		return nullWrpcId, err
	}
	return &wrpc.Id{Id: created_id}, err
}

// Util func to call server side delete func passing in Id and returning error, if any
func serverCallDelete(in string, call func(string) error) (*wrpc.Text, error) {
	err := call(in)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return nullWrpcText, err
}

// Call server side DeleteCollection
func (s *grpcServer) DeleteCollection(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return serverCallDelete(in.Id, s.handler.DeleteCollection)
}

// Call server side DeleteItem
func (s *grpcServer) DeleteItem(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return serverCallDelete(in.Id, s.handler.DeleteItem)
}

// Call server side DeleteVersion
func (s *grpcServer) DeleteVersion(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return serverCallDelete(in.Id, s.handler.DeleteVersion)
}

// Call server side DeleteResource
func (s *grpcServer) DeleteResource(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return serverCallDelete(in.Id, s.handler.DeleteResource)
}

// Util func to convert rpc QueryDesc objects to wysteria native QueryDesc objects
func convRQueryDescs(in ...*wrpc.QueryDesc) []*wyc.QueryDesc {
	result := []*wyc.QueryDesc{}
	for _, q := range in {
		result = append(
			result,
			&wyc.QueryDesc{
				Parent:        q.Parent,
				Id:            q.Id,
				VersionNumber: q.VersionNumber,
				ItemType:      q.ItemType,
				Variant:       q.Variant,
				Facets:        q.Facets,
				Name:          q.Name,
				ResourceType:  q.ResourceType,
				Location:      q.Location,
				LinkSrc:       q.LinkSrc,
				LinkDst:       q.LinkDst,
			},
		)
	}
	return result
}

// Util func to convert wysteria native Collection objects to rpc Collection objects
func convWCollections(in ...*wyc.Collection) *wrpc.Collections {
	result := []*wrpc.Collection{}
	for _, i := range in {
		result = append(
			result,
			&wrpc.Collection{
				Id:   i.Id,
				Name: i.Name,
			},
		)
	}
	return &wrpc.Collections{
		All: result,
	}
}

// Call server side FindCollections passing in given query, return results
func (s *grpcServer) FindCollections(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Collections, error) {
	results, err := s.handler.FindCollections(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}
	return convWCollections(results...), nil
}

// Call server side FindItems passing in given query, return results
func (s *grpcServer) FindItems(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Items, error) {
	results, err := s.handler.FindItems(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}

	res := &wrpc.Items{All: []*wrpc.Item{}}
	for _, i := range results {
		res.All = append(res.All, convWItem(i))
	}
	return res, nil
}

// Call server side FindVersions passing in given query, return results
func (s *grpcServer) FindVersions(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Versions, error) {
	results, err := s.handler.FindVersions(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}

	res := &wrpc.Versions{All: []*wrpc.Version{}}
	for _, i := range results {
		res.All = append(res.All, convWVersion(i))
	}
	return res, nil
}

// Call server side FindResources passing in given query, return results
func (s *grpcServer) FindResources(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Resources, error) {
	results, err := s.handler.FindResources(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}

	res := &wrpc.Resources{All: []*wrpc.Resource{}}
	for _, i := range results {
		res.All = append(res.All, convWResource(i))
	}
	return res, nil
}

// Call server side FindLinks passing in given query, return results
func (s *grpcServer) FindLinks(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Links, error) {
	results, err := s.handler.FindLinks(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}

	res := &wrpc.Links{All: []*wrpc.Link{}}
	for _, i := range results {
		res.All = append(res.All, convWLink(i))
	}
	return res, nil
}

// Call server side GetPublishedVersion, pass in given Item Id and return results
func (s *grpcServer) GetPublishedVersion(_ context.Context, in *wrpc.Id) (*wrpc.Version, error) {
	version, err := s.handler.GetPublishedVersion(in.Id)
	if err != nil {
		return nil, err
	}
	return convWVersion(version), err
}

// Call server side PublishVersion passing in Version id, return results
func (s *grpcServer) PublishVersion(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	err := s.handler.PublishVersion(in.Id)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return nullWrpcText, err
}
