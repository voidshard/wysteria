package middleware

import (
	wyc "github.com/voidshard/wysteria/common"
	wrpc "github.com/voidshard/wysteria/common/middleware/wysteria_grpc"
	"google.golang.org/grpc"
	"golang.org/x/net/context"
	"errors"
	"net"
	"google.golang.org/grpc/reflection"
	"log"
)

type grpcClient struct {
	config string
	conn *grpc.ClientConn
	client wrpc.WysteriaGrpcClient
}

func (c *grpcClient) Connect(config string) error {
	conn, err := grpc.Dial(config, grpc.WithInsecure())
	if err != nil {
		return err
	}
	c.conn = conn

	c.client = wrpc.NewWysteriaGrpcClient(c.conn)
	return nil
}

func (c *grpcClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func (c *grpcClient) CreateCollection(name string) (string, error) {
	result, err := c.client.CreateCollection(
		context.Background(),
		&wrpc.Text{Text: name},
	)

	if err != nil {
		return "", err
	}
	if result.Error.Text != "" {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

func convWItem (in *wyc.Item) *wrpc.Item {
	return &wrpc.Item{
		Id: in.Id,
		Parent: in.Parent,
		ItemType: in.ItemType,
		Variant: in.Variant,
		Links: in.Links,
		Facets: in.Facets,
	}
}

func (c *grpcClient) CreateItem(in *wyc.Item) (string, error) {
	result, err := c.client.CreateItem(
		context.Background(),
		convWItem(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error.Text != "" {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

func convWVersion (in *wyc.Version) *wrpc.Version {
	return &wrpc.Version{
		Id: in.Id,
		Parent: in.Parent,
		Number: in.Number,
		Links: in.Links,
		Facets: in.Facets,
		Resources: in.Resources,
	}
}

func (c *grpcClient) CreateVersion(in *wyc.Version) (string, error) {
	result, err := c.client.CreateVersion(
		context.Background(),
		convWVersion(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error.Text != "" {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

func convWResource (in *wyc.Resource) *wrpc.Resource {
	return &wrpc.Resource{
		Id: in.Id,
		Parent: in.Parent,
		Name: in.Name,
		Location: in.Location,
		ResourceType: in.ResourceType,
	}
}


func (c *grpcClient) CreateResource(in *wyc.Resource) (string, error) {
	result, err := c.client.CreateResource(
		context.Background(),
		convWResource(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error.Text != "" {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

func convWLink (in *wyc.Link) *wrpc.Link {
	return &wrpc.Link{
		Id: in.Id,
		Src: in.Src,
		Name: in.Name,
		Dst: in.Dst,
	}
}

func (c *grpcClient) CreateLink(in *wyc.Link) (string, error) {
	result, err := c.client.CreateLink(
		context.Background(),
		convWLink(in),
	)

	if err != nil {
		return "", err
	}
	if result.Error.Text != "" {
		return "", errors.New(result.Error.Text)
	}
	return result.Id, nil
}

func client_call_delete(delete_id string, call func (ctx context.Context, in *wrpc.Id, opts ...grpc.CallOption) (*wrpc.Text, error) ) error {
	result, err := call(context.Background(), &wrpc.Id{Id: delete_id})
	if err != nil {
		return err
	}
	if result.Text != "" {
		return errors.New(result.Text)
	}
	return nil
}

func (c *grpcClient) DeleteCollection(in string) error {
	return client_call_delete(in, c.client.DeleteCollection)
}

func (c *grpcClient) DeleteItem(in string) error {
	return client_call_delete(in, c.client.DeleteItem)
}

func (c *grpcClient) DeleteVersion(in string) error {
	return client_call_delete(in, c.client.DeleteVersion)
}

func (c *grpcClient) DeleteResource(in string) error {
	return client_call_delete(in, c.client.DeleteResource)
}

func convWQueryDescs(in ...*wyc.QueryDesc) *wrpc.QueryDescs {
	result := []*wrpc.QueryDesc{}
	for _, q := range in {
		result = append(
			result,
			&wrpc.QueryDesc{
				Parent: q.Parent,
				Id: q.Id,
				VersionNumber: q.VersionNumber,
				ItemType: q.ItemType,
				Variant: q.Variant,
				Facets: q.Facets,
				Name: q.Name,
				ResourceType: q.ResourceType,
				Location: q.Location,
				LinkSrc: q.LinkSrc,
				LinkDst: q.LinkDst,
			},
		)
	}
	return &wrpc.QueryDescs{All: result}
}

func convRCollections(in ...*wrpc.Collection) []*wyc.Collection {
	result := []*wyc.Collection{}
	for _, i := range in {
		result = append(result, &wyc.Collection{
			Id: i.Id,
			Name: i.Name,
		})
	}
	return result
}

func (c *grpcClient) FindCollections(in []*wyc.QueryDesc) ([]*wyc.Collection, error) {
	result, err := c.client.FindCollections(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRCollections(result.All...), nil
}

func convRItems(in ...*wrpc.Item) []*wyc.Item {
	result := []*wyc.Item{}
	for _, i := range in {
		result = append(result, &wyc.Item{
			Id: i.Id,
			Parent: i.Parent,
			ItemType: i.ItemType,
			Variant: i.Variant,
			Links: i.Links,
			Facets: i.Facets,
		})
	}
	return result
}

func (c *grpcClient) FindItems(in []*wyc.QueryDesc) ([]*wyc.Item, error) {
	result, err := c.client.FindItems(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRItems(result.All...), nil
}

func convRVersions(in ...*wrpc.Version) []*wyc.Version {
	result := []*wyc.Version{}
	for _, i := range in {
		result = append(result, &wyc.Version{
			Id: i.Id,
			Parent: i.Parent,
			Links: i.Links,
			Facets: i.Facets,
			Number: i.Number,
			Resources: i.Resources,
		})
	}
	return result
}

func (c *grpcClient) FindVersions(in []*wyc.QueryDesc) ([]*wyc.Version, error) {
	result, err := c.client.FindVersions(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRVersions(result.All...), nil
}

func convRResources(in ...*wrpc.Resource) []*wyc.Resource {
	result := []*wyc.Resource{}
	for _, i := range in {
		result = append(result, &wyc.Resource{
			Id: i.Id,
			Parent: i.Parent,
			Name: i.Name,
			ResourceType: i.ResourceType,
			Location: i.Location,
		})
	}
	return result
}

func (c *grpcClient) FindResources(in []*wyc.QueryDesc) ([]*wyc.Resource, error) {
	result, err := c.client.FindResources(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRResources(result.All...), nil	
}

func convRLinks(in ...*wrpc.Link) []*wyc.Link {
	result := []*wyc.Link{}
	for _, i := range in {
		result = append(result, &wyc.Link{
			Id: i.Id,
			Name: i.Name,
			Src: i.Src,
			Dst: i.Dst,
		})
	}
	return result
}

func (c *grpcClient) FindLinks(in []*wyc.QueryDesc) ([]*wyc.Link, error) {
	result, err := c.client.FindLinks(context.Background(), convWQueryDescs(in...))
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRLinks(result.All...), nil
}

func convRVersion(in *wrpc.Version) *wyc.Version {
	return &wyc.Version{
		Id: in.Id,
		Parent: in.Parent,
		Links: in.Links,
		Facets: in.Facets,
		Number: in.Number,
		Resources: in.Resources,
	}
}

func (c *grpcClient) GetPublishedVersion(in string) (*wyc.Version, error) {
	result, err := c.client.GetPublishedVersion(context.Background(), &wrpc.Id{Id: in})
	if err != nil {
		return nil, err
	}

	if result.Error.Text != "" {
		return nil, errors.New(result.Error.Text)
	}
	return convRVersion(result), nil
}

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

func newGrpcClient() EndpointClient {
	return &grpcClient{}
}

type grpcServer struct {
	conn net.Listener
	config string
	server wrpc.WysteriaGrpcServer
	handler ServerHandler
}

func (s *grpcServer) Initialize(config string, handler *ServerHandler) error {
	conn, err := net.Listen("tcp", config)
	if err != nil {
		return err
	}

	s.conn = conn
	s.handler = handler

	rpc_server := grpc.NewServer()
	wrpc.RegisterWysteriaGrpcServer(
		rpc_server,
		s,
	)

	// Register reflection service on gRPC server.
	reflection.Register(rpc_server)

	go func() {
		log.Fatal(rpc_server.Serve(s.conn))
	}()
}

func (s *grpcServer) CreateCollection(_ context.Context, in *wrpc.Text) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateCollection(in.Text)
	if err != nil {
		return &wrpc.Id{Id: ""}, err
	}
	return &wrpc.Id{Id: created_id}, err
}

func (s *grpcServer) CreateItem(_ context.Context, in *wrpc.Item) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateItem(convRItems(in)[0])
	if err != nil {
		return &wrpc.Id{Id: ""}, err
	}
	return &wrpc.Id{Id: created_id}, err
}

func (s *grpcServer) CreateVersion(_ context.Context, in *wrpc.Version) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateVersion(convRVersion(in))
	if err != nil {
		return &wrpc.Id{Id: ""}, err
	}
	return &wrpc.Id{Id: created_id}, err
}

func (s *grpcServer) CreateResource(_ context.Context, in *wrpc.Resource) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateResource(convRResources(in)[0])
	if err != nil {
		return &wrpc.Id{Id: ""}, err
	}
	return &wrpc.Id{Id: created_id}, err
}

func (s *grpcServer) CreateLink(_ context.Context, in *wrpc.Link) (*wrpc.Id, error) {
	created_id, err := s.handler.CreateLink(convRLinks(in)[0])
	if err != nil {
		return &wrpc.Id{Id: ""}, err
	}
	return &wrpc.Id{Id: created_id}, err
}

func server_call_delete(in string, call func(string) error) (*wrpc.Text, error) {
	err := call(in)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return &wrpc.Text{Text: ""}, err
}

func (s *grpcServer) DeleteCollection(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return server_call_delete(in.Id, s.handler.DeleteCollection)
}

func (s *grpcServer) DeleteItem(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return server_call_delete(in.Id, s.handler.DeleteItem)
}
func (s *grpcServer) DeleteVersion(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return server_call_delete(in.Id, s.handler.DeleteVersion)
}

func (s *grpcServer) DeleteResource(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	return server_call_delete(in.Id, s.handler.DeleteResource)
}

func convRQueryDescs(in ...*wrpc.QueryDesc) []*wyc.QueryDesc {
	result := []*wyc.QueryDesc{}
	for _, q := range in {
		result = append(
			result,
			&wyc.QueryDesc{
				Parent: q.Parent,
				Id: q.Id,
				VersionNumber: q.VersionNumber,
				ItemType: q.ItemType,
				Variant: q.Variant,
				Facets: q.Facets,
				Name: q.Name,
				ResourceType: q.ResourceType,
				Location: q.Location,
				LinkSrc: q.LinkSrc,
				LinkDst: q.LinkDst,
			},
		)
	}
	return result
}

func convWCollections (in ...*wyc.Collection) *wrpc.Collections {
	result := []*wrpc.Collection{}
	for _, i := range in {
		result = append(
			result,
			&wrpc.Collection{
				Id: i.Id,
				Name: i.Name,
			},
		)
	}
	return &wrpc.Collections{
		All: result,
	}
}

func (s *grpcServer) FindCollections(_ context.Context, in *wrpc.QueryDescs) (*wrpc.Collections, error) {
	results, err := s.handler.FindCollections(convRQueryDescs(in.All...))
	if err != nil {
		return nil, err
	}
	return convWCollections(results...), nil
}

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

func (s *grpcServer) GetPublishedVersion(_ context.Context, in *wrpc.Id) (*wrpc.Version, error) {
	version, err := s.handler.GetPublishedVersion(in.Id)
	if err != nil {
		return nil, err
	}
	return convWVersion(version), err
}

func (s *grpcServer) PublishVersion(_ context.Context, in *wrpc.Id) (*wrpc.Text, error) {
	err := s.handler.PublishVersion(in.Id)
	if err != nil {
		return &wrpc.Text{Text: err.Error()}, err
	}
	return &wrpc.Text{Text: ""}, err
}

func (s *grpcServer) Shutdown() error {
	if s.server == nil {
		return nil
	}
	s.conn.Close()
}

func newGrpcServer(config string) EndpointServer {
	return &grpcServer{config: config}
}
