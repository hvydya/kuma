// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mesh/v1alpha1/mux.proto

package v1alpha1

import (
	context "context"
	fmt "fmt"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	proto "github.com/golang/protobuf/proto"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Message struct {
	// Types that are valid to be assigned to Value:
	//	*Message_Request
	//	*Message_Response
	Value                isMessage_Value `protobuf_oneof:"value"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *Message) Reset()         { *m = Message{} }
func (m *Message) String() string { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()    {}
func (*Message) Descriptor() ([]byte, []int) {
	return fileDescriptor_df76defa729b08eb, []int{0}
}

func (m *Message) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Message.Unmarshal(m, b)
}
func (m *Message) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Message.Marshal(b, m, deterministic)
}
func (m *Message) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Message.Merge(m, src)
}
func (m *Message) XXX_Size() int {
	return xxx_messageInfo_Message.Size(m)
}
func (m *Message) XXX_DiscardUnknown() {
	xxx_messageInfo_Message.DiscardUnknown(m)
}

var xxx_messageInfo_Message proto.InternalMessageInfo

type isMessage_Value interface {
	isMessage_Value()
}

type Message_Request struct {
	Request *v2.DiscoveryRequest `protobuf:"bytes,1,opt,name=request,proto3,oneof"`
}

type Message_Response struct {
	Response *v2.DiscoveryResponse `protobuf:"bytes,2,opt,name=response,proto3,oneof"`
}

func (*Message_Request) isMessage_Value() {}

func (*Message_Response) isMessage_Value() {}

func (m *Message) GetValue() isMessage_Value {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Message) GetRequest() *v2.DiscoveryRequest {
	if x, ok := m.GetValue().(*Message_Request); ok {
		return x.Request
	}
	return nil
}

func (m *Message) GetResponse() *v2.DiscoveryResponse {
	if x, ok := m.GetValue().(*Message_Response); ok {
		return x.Response
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Message) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Message_Request)(nil),
		(*Message_Response)(nil),
	}
}

func init() {
	proto.RegisterType((*Message)(nil), "kuma.mesh.v1alpha1.Message")
}

func init() { proto.RegisterFile("mesh/v1alpha1/mux.proto", fileDescriptor_df76defa729b08eb) }

var fileDescriptor_df76defa729b08eb = []byte{
	// 247 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0xd0, 0x3f, 0x4b, 0xc3, 0x40,
	0x18, 0x06, 0xf0, 0x46, 0xd0, 0xc8, 0x89, 0x20, 0xb7, 0x58, 0xaa, 0xa8, 0x38, 0x05, 0x87, 0xf7,
	0x6c, 0xdc, 0x04, 0x97, 0xe2, 0xd0, 0x25, 0x4b, 0xba, 0xb9, 0xbd, 0x8d, 0x2f, 0xcd, 0x61, 0xae,
	0x77, 0xbd, 0x7f, 0xb4, 0x1f, 0xc2, 0xef, 0x2c, 0xbd, 0x18, 0x41, 0xc4, 0x4e, 0x37, 0x3c, 0xcf,
	0x0f, 0xee, 0x7d, 0xd8, 0xa5, 0x22, 0xd7, 0x8a, 0x38, 0xc5, 0xce, 0xb4, 0x38, 0x15, 0x2a, 0x6c,
	0xc1, 0x58, 0xed, 0x35, 0xe7, 0x1f, 0x41, 0x21, 0xec, 0x53, 0x18, 0xd2, 0xc9, 0x35, 0xad, 0xa3,
	0xde, 0x09, 0x34, 0x52, 0xc4, 0x52, 0xbc, 0x4b, 0xd7, 0xe8, 0x48, 0x76, 0xd7, 0x8b, 0xfb, 0xcf,
	0x8c, 0xe5, 0x15, 0x39, 0x87, 0x2b, 0xe2, 0xcf, 0x2c, 0xb7, 0xb4, 0x09, 0xe4, 0xfc, 0x38, 0xbb,
	0xcb, 0x8a, 0xb3, 0xf2, 0x06, 0x92, 0x05, 0x34, 0x12, 0x62, 0x09, 0xaf, 0x83, 0xad, 0xfb, 0xd6,
	0x7c, 0x54, 0x0f, 0x80, 0xbf, 0xb0, 0x53, 0x4b, 0xce, 0xe8, 0xb5, 0xa3, 0xf1, 0x51, 0xc2, 0xb7,
	0xff, 0xe2, 0xbe, 0x36, 0x1f, 0xd5, 0x3f, 0x64, 0x96, 0xb3, 0xe3, 0x88, 0x5d, 0xa0, 0x12, 0xd9,
	0x45, 0x15, 0x3a, 0x2f, 0x4d, 0x47, 0xdb, 0x05, 0xd9, 0x28, 0x1b, 0xe2, 0x15, 0x3b, 0x5f, 0x78,
	0x4b, 0xa8, 0x86, 0x8f, 0x5e, 0xc1, 0xdf, 0x3b, 0xe1, 0x3b, 0x9c, 0x1c, 0x0a, 0x8b, 0xec, 0x31,
	0x9b, 0x3d, 0xbc, 0x15, 0x2b, 0xe9, 0xdb, 0xb0, 0x84, 0x46, 0x2b, 0xb1, 0x2f, 0xb7, 0x9b, 0xf4,
	0xa4, 0x8d, 0x7e, 0x4d, 0xbb, 0x3c, 0x49, 0x2b, 0x3d, 0x7d, 0x05, 0x00, 0x00, 0xff, 0xff, 0x28,
	0x7d, 0x51, 0xc2, 0x72, 0x01, 0x00, 0x00,
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MultiplexServiceClient is the client API for MultiplexService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MultiplexServiceClient interface {
	StreamMessage(ctx context.Context, opts ...grpc.CallOption) (MultiplexService_StreamMessageClient, error)
}

type multiplexServiceClient struct {
	cc *grpc.ClientConn
}

func NewMultiplexServiceClient(cc *grpc.ClientConn) MultiplexServiceClient {
	return &multiplexServiceClient{cc}
}

func (c *multiplexServiceClient) StreamMessage(ctx context.Context, opts ...grpc.CallOption) (MultiplexService_StreamMessageClient, error) {
	stream, err := c.cc.NewStream(ctx, &_MultiplexService_serviceDesc.Streams[0], "/kuma.mesh.v1alpha1.MultiplexService/StreamMessage", opts...)
	if err != nil {
		return nil, err
	}
	x := &multiplexServiceStreamMessageClient{stream}
	return x, nil
}

type MultiplexService_StreamMessageClient interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type multiplexServiceStreamMessageClient struct {
	grpc.ClientStream
}

func (x *multiplexServiceStreamMessageClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *multiplexServiceStreamMessageClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// MultiplexServiceServer is the server API for MultiplexService service.
type MultiplexServiceServer interface {
	StreamMessage(MultiplexService_StreamMessageServer) error
}

// UnimplementedMultiplexServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMultiplexServiceServer struct {
}

func (*UnimplementedMultiplexServiceServer) StreamMessage(srv MultiplexService_StreamMessageServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamMessage not implemented")
}

func RegisterMultiplexServiceServer(s *grpc.Server, srv MultiplexServiceServer) {
	s.RegisterService(&_MultiplexService_serviceDesc, srv)
}

func _MultiplexService_StreamMessage_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MultiplexServiceServer).StreamMessage(&multiplexServiceStreamMessageServer{stream})
}

type MultiplexService_StreamMessageServer interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type multiplexServiceStreamMessageServer struct {
	grpc.ServerStream
}

func (x *multiplexServiceStreamMessageServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *multiplexServiceStreamMessageServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _MultiplexService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "kuma.mesh.v1alpha1.MultiplexService",
	HandlerType: (*MultiplexServiceServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMessage",
			Handler:       _MultiplexService_StreamMessage_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "mesh/v1alpha1/mux.proto",
}
