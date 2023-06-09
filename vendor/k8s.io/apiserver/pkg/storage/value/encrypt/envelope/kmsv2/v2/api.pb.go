/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: api.proto

package v2

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
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
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// EncryptedObject is the representation of data stored in etcd after envelope encryption.
type EncryptedObject struct {
	// EncryptedData is the encrypted data.
	EncryptedData []byte `protobuf:"bytes,1,opt,name=encryptedData,proto3" json:"encryptedData,omitempty"`
	// KeyID is the KMS key ID used for encryption operations.
	KeyID string `protobuf:"bytes,2,opt,name=keyID,proto3" json:"keyID,omitempty"`
	// EncryptedDEK is the encrypted DEK.
	EncryptedDEK []byte `protobuf:"bytes,3,opt,name=encryptedDEK,proto3" json:"encryptedDEK,omitempty"`
	// Annotations is additional metadata that was provided by the KMS plugin.
	Annotations          map[string][]byte `protobuf:"bytes,4,rep,name=annotations,proto3" json:"annotations,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	XXX_NoUnkeyedLiteral struct{}          `json:"-"`
	XXX_unrecognized     []byte            `json:"-"`
	XXX_sizecache        int32             `json:"-"`
}

func (m *EncryptedObject) Reset()         { *m = EncryptedObject{} }
func (m *EncryptedObject) String() string { return proto.CompactTextString(m) }
func (*EncryptedObject) ProtoMessage()    {}
func (*EncryptedObject) Descriptor() ([]byte, []int) {
	return fileDescriptor_00212fb1f9d3bf1c, []int{0}
}
func (m *EncryptedObject) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_EncryptedObject.Unmarshal(m, b)
}
func (m *EncryptedObject) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_EncryptedObject.Marshal(b, m, deterministic)
}
func (m *EncryptedObject) XXX_Merge(src proto.Message) {
	xxx_messageInfo_EncryptedObject.Merge(m, src)
}
func (m *EncryptedObject) XXX_Size() int {
	return xxx_messageInfo_EncryptedObject.Size(m)
}
func (m *EncryptedObject) XXX_DiscardUnknown() {
	xxx_messageInfo_EncryptedObject.DiscardUnknown(m)
}

var xxx_messageInfo_EncryptedObject proto.InternalMessageInfo

func (m *EncryptedObject) GetEncryptedData() []byte {
	if m != nil {
		return m.EncryptedData
	}
	return nil
}

func (m *EncryptedObject) GetKeyID() string {
	if m != nil {
		return m.KeyID
	}
	return ""
}

func (m *EncryptedObject) GetEncryptedDEK() []byte {
	if m != nil {
		return m.EncryptedDEK
	}
	return nil
}

func (m *EncryptedObject) GetAnnotations() map[string][]byte {
	if m != nil {
		return m.Annotations
	}
	return nil
}

func init() {
	proto.RegisterType((*EncryptedObject)(nil), "v2.EncryptedObject")
	proto.RegisterMapType((map[string][]byte)(nil), "v2.EncryptedObject.AnnotationsEntry")
}

func init() { proto.RegisterFile("api.proto", fileDescriptor_00212fb1f9d3bf1c) }

var fileDescriptor_00212fb1f9d3bf1c = []byte{
	// 244 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0xb1, 0x4b, 0x03, 0x31,
	0x14, 0xc6, 0xc9, 0x9d, 0x0a, 0x97, 0x9e, 0x58, 0x82, 0xc3, 0xe1, 0x74, 0x94, 0x0e, 0x37, 0x25,
	0x10, 0x97, 0x22, 0x52, 0x50, 0x7a, 0x82, 0x38, 0x08, 0x19, 0xdd, 0xd2, 0xfa, 0x28, 0x67, 0x6a,
	0x12, 0x92, 0x18, 0xc8, 0x9f, 0xee, 0x26, 0x4d, 0x95, 0xda, 0xdb, 0xde, 0xf7, 0xf1, 0xfb, 0xe0,
	0xc7, 0xc3, 0x95, 0xb4, 0x03, 0xb5, 0xce, 0x04, 0x43, 0x8a, 0xc8, 0x67, 0xdf, 0x08, 0x5f, 0xf5,
	0x7a, 0xe3, 0x92, 0x0d, 0xf0, 0xfe, 0xba, 0xfe, 0x80, 0x4d, 0x20, 0x73, 0x7c, 0x09, 0x7f, 0xd5,
	0x4a, 0x06, 0xd9, 0xa0, 0x16, 0x75, 0xb5, 0x38, 0x2d, 0xc9, 0x35, 0x3e, 0x57, 0x90, 0x9e, 0x57,
	0x4d, 0xd1, 0xa2, 0xae, 0x12, 0x87, 0x40, 0x66, 0xb8, 0x3e, 0x62, 0xfd, 0x4b, 0x53, 0xe6, 0xe9,
	0x49, 0x47, 0x9e, 0xf0, 0x44, 0x6a, 0x6d, 0x82, 0x0c, 0x83, 0xd1, 0xbe, 0x39, 0x6b, 0xcb, 0x6e,
	0xc2, 0xe7, 0x34, 0x72, 0x3a, 0x32, 0xa1, 0x0f, 0x47, 0xac, 0xd7, 0xc1, 0x25, 0xf1, 0x7f, 0x78,
	0xb3, 0xc4, 0xd3, 0x31, 0x40, 0xa6, 0xb8, 0x54, 0x90, 0xb2, 0x71, 0x25, 0xf6, 0xe7, 0xde, 0x33,
	0xca, 0xdd, 0x17, 0x64, 0xcf, 0x5a, 0x1c, 0xc2, 0x5d, 0xb1, 0x40, 0x8f, 0xcb, 0xb7, 0x7b, 0xb5,
	0xf0, 0x74, 0x30, 0x4c, 0xda, 0xc1, 0x83, 0x8b, 0xe0, 0x98, 0x55, 0x5b, 0xe6, 0x83, 0x71, 0x72,
	0x0b, 0x2c, 0x93, 0xec, 0x57, 0x9d, 0x81, 0x8e, 0xb0, 0x33, 0x16, 0x98, 0xfa, 0xf4, 0x91, 0xb3,
	0xc8, 0xd7, 0x17, 0xf9, 0x8d, 0xb7, 0x3f, 0x01, 0x00, 0x00, 0xff, 0xff, 0x00, 0x80, 0x43, 0x93,
	0x53, 0x01, 0x00, 0x00,
}
