package gen_proto

var normalClientFunc = `
func (c *${serviceName}Client) ${methodName}(ctx context.Context, in *${req}, opts ...grpc.CallOption) (*${resp}, error) {
	out := new(${resp})
	err := c.cc.Invoke(ctx, c.name, "${methodPath}", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}`
var allStreamClient = `
type ${serviceName}_${methodName}Client struct {
	grpcx.ClientStream
}

func (x *${serviceName}_${methodName}Client) Send(m *${req}) error {
	return x.ClientStream.Send(m)
}

func (x *${serviceName}_${methodName}Client) Recv() (*${resp}, error) {
	m := new(${resp})
	if err := x.ClientStream.Recv(m); err != nil {
		return nil, err
	}
	return m, nil
}`
var allStreamClientFunc = `
func (c *${serviceName}Client) ${methodName}(ctx context.Context, opts ...grpc.CallOption) (*${serviceName}_${methodName}Client, error) {
	stream, err := c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName: "${methodName}", ServerStreams: true, ClientStreams: true}, c.name, "${methodPath}", opts...)
	if err != nil {
		return nil, err
	}
	x := &${serviceName}_${methodName}Client{stream}
	return x, nil
}`

var clientStreamClient = `
type ${serviceName}_${methodName}Client struct {
	grpcx.ClientStream
}

func (x *${serviceName}_${methodName}Client) Send(m *${req}) error {
	return x.ClientStream.Send(m)
}`

var clientStreamClientFunc = `
func (c *${serviceName}Client) ${methodName}(ctx context.Context, opts ...grpc.CallOption) (*${serviceName}_${methodName}Client, error) {
	stream, err := c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName: "${methodName}", ClientStreams: true}, c.name, "${methodPath}", opts...)
	if err != nil {
		return nil, err
	}
	x := &${serviceName}_${methodName}Client{stream}
	return x, nil
}`

var serverStreamClient = `
type ${serviceName}_${methodName}Client struct {
	grpcx.ClientStream
}

func (x *${serviceName}_${methodName}Client) Recv() (*${resp}, error) {
	m := new(${resp})
	if err := x.ClientStream.Recv(m); err != nil {
		return nil, err
	}
	return m, nil
}`

var serverStreamClientFunc = `
func (c *${serviceName}Client) ${methodName}(ctx context.Context, in *${req}, opts ...grpc.CallOption) (*${serviceName}_${methodName}Client, error) {
	stream, err := c.cc.NewStream(ctx, &grpc.StreamDesc{
		StreamName: "${methodName}", ServerStreams: true}, c.name, "${methodPath}", opts...)
	if err != nil {
		return nil, err
	}
	x := &${serviceName}_${methodName}Client{stream}
	if err := x.ClientStream.Send(in); err != nil {
		return nil, err
	}
	return x, nil
}`

var allStreamServer = `
type ${serviceName}_${methodName}Server struct {
	grpc.ServerStream
}

func (x *${serviceName}_${methodName}Server) SetStream(s grpc.ServerStream) {
	x.ServerStream = s
}

func (x *${serviceName}_${methodName}Server) Send(m *${resp}) error {
	return x.ServerStream.SendMsg(m)
}

func (x *${serviceName}_${methodName}Server) Recv() (*${req}, error) {
	m := new(${req})
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}`

var allStreamServerFunc = `
func (the *${serviceName}) ${methodName}(ctx context.Context, s *pb.${serviceName}_${methodName}Server) error {
	return nil
}`

var serverStreamServer = `
type ${serviceName}_${methodName}Server struct {
	grpc.ServerStream
}

func (x *${serviceName}_${methodName}Server) SetStream(s grpc.ServerStream) {
	x.ServerStream = s
}

func (x *${serviceName}_${methodName}Server) Send(m *${resp}) error {
	return x.ServerStream.SendMsg(m)
}`

var serverStreamServerFunc = `
func (the *${serviceName}) ${methodName}(ctx context.Context, req *pb.${req}, s *pb.${serviceName}_${methodName}Server) error {
	return nil
}`

var clientStreamServer = `
type ${serviceName}_${methodName}Server struct {
	grpc.ServerStream
}

func (x *${serviceName}_${methodName}Server) SetStream(s grpc.ServerStream) {
	x.ServerStream = s
}

func (x *${serviceName}_${methodName}Server) Recv() (*${req}, error) {
	m := new(${req})
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}`

var clientStreamServerFunc = `
func (the *${serviceName}) ${methodName}(ctx context.Context, s *pb.${serviceName}_${methodName}Server) error {
	return nil
}`

var nolmalServer = `
func (the *${serviceName}) ${methodName}(ctx context.Context, req *pb.${req}, resp *pb.${resp}) error {
	return nil
}`
