// Code generated by protoc-gen-go. DO NOT EDIT.
// source: mesh/v1alpha1/selector.proto

package v1alpha1

import (
	fmt "fmt"
	proto "github.com/golang/protobuf/proto"
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

// Selector defines structure for selecting tags for given dataplane
type Selector struct {
	// Tags to match, can be used for both source and destinations
	Match                map[string]string `protobuf:"bytes,1,rep,name=match,proto3" json:"match,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *Selector) Reset()         { *m = Selector{} }
func (m *Selector) String() string { return proto.CompactTextString(m) }
func (*Selector) ProtoMessage()    {}
func (*Selector) Descriptor() ([]byte, []int) {
	return fileDescriptor_95def39c1d383442, []int{0}
}

func (m *Selector) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Selector.Unmarshal(m, b)
}
func (m *Selector) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Selector.Marshal(b, m, deterministic)
}
func (m *Selector) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Selector.Merge(m, src)
}
func (m *Selector) XXX_Size() int {
	return xxx_messageInfo_Selector.Size(m)
}
func (m *Selector) XXX_DiscardUnknown() {
	xxx_messageInfo_Selector.DiscardUnknown(m)
}

var xxx_messageInfo_Selector proto.InternalMessageInfo

func (m *Selector) GetMatch() map[string]string {
	if m != nil {
		return m.Match
	}
	return nil
}

func init() {
	proto.RegisterType((*Selector)(nil), "kuma.mesh.v1alpha1.Selector")
	proto.RegisterMapType((map[string]string)(nil), "kuma.mesh.v1alpha1.Selector.MatchEntry")
}

func init() { proto.RegisterFile("mesh/v1alpha1/selector.proto", fileDescriptor_95def39c1d383442) }

var fileDescriptor_95def39c1d383442 = []byte{
	// 183 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x92, 0xc9, 0x4d, 0x2d, 0xce,
	0xd0, 0x2f, 0x33, 0x4c, 0xcc, 0x29, 0xc8, 0x48, 0x34, 0xd4, 0x2f, 0x4e, 0xcd, 0x49, 0x4d, 0x2e,
	0xc9, 0x2f, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x12, 0xca, 0x2e, 0xcd, 0x4d, 0xd4, 0x03,
	0x29, 0xd1, 0x83, 0x29, 0x51, 0x6a, 0x66, 0xe4, 0xe2, 0x08, 0x86, 0x2a, 0x13, 0xb2, 0xe5, 0x62,
	0xcd, 0x4d, 0x2c, 0x49, 0xce, 0x90, 0x60, 0x54, 0x60, 0xd6, 0xe0, 0x36, 0x52, 0xd7, 0xc3, 0xd4,
	0xa0, 0x07, 0x53, 0xac, 0xe7, 0x0b, 0x52, 0xe9, 0x9a, 0x57, 0x52, 0x54, 0x19, 0x04, 0xd1, 0x25,
	0x65, 0xc1, 0xc5, 0x85, 0x10, 0x14, 0x12, 0xe0, 0x62, 0xce, 0x4e, 0xad, 0x94, 0x60, 0x54, 0x60,
	0xd4, 0xe0, 0x0c, 0x02, 0x31, 0x85, 0x44, 0xb8, 0x58, 0xcb, 0x12, 0x73, 0x4a, 0x53, 0x25, 0x98,
	0xc0, 0x62, 0x10, 0x8e, 0x15, 0x93, 0x05, 0xa3, 0x93, 0x56, 0x94, 0x46, 0x7a, 0x66, 0x49, 0x46,
	0x69, 0x92, 0x5e, 0x72, 0x7e, 0xae, 0x3e, 0xc8, 0xd6, 0x8c, 0x42, 0x30, 0xa5, 0x9f, 0x58, 0x90,
	0xa9, 0x8f, 0xe2, 0xa9, 0x24, 0x36, 0xb0, 0x67, 0x8c, 0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x99,
	0x36, 0x1b, 0x88, 0xec, 0x00, 0x00, 0x00,
}
