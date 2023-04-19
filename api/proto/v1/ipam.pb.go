// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.11.4
// source: ipam.proto

package proto_v1

import (
	_ "google.golang.org/genproto/googleapis/api/annotations"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type IPAMRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Command   string `protobuf:"bytes,1,opt,name=command,proto3" json:"command,omitempty"`
	Id        string `protobuf:"bytes,2,opt,name=id,proto3" json:"id,omitempty"`
	IfName    string `protobuf:"bytes,3,opt,name=ifName,proto3" json:"ifName,omitempty"`
	Namespace string `protobuf:"bytes,4,opt,name=namespace,proto3" json:"namespace,omitempty"`
	Name      string `protobuf:"bytes,5,opt,name=name,proto3" json:"name,omitempty"`
	Uid       string `protobuf:"bytes,6,opt,name=uid,proto3" json:"uid,omitempty"`
}

func (x *IPAMRequest) Reset() {
	*x = IPAMRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipam_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPAMRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPAMRequest) ProtoMessage() {}

func (x *IPAMRequest) ProtoReflect() protoreflect.Message {
	mi := &file_ipam_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPAMRequest.ProtoReflect.Descriptor instead.
func (*IPAMRequest) Descriptor() ([]byte, []int) {
	return file_ipam_proto_rawDescGZIP(), []int{0}
}

func (x *IPAMRequest) GetCommand() string {
	if x != nil {
		return x.Command
	}
	return ""
}

func (x *IPAMRequest) GetId() string {
	if x != nil {
		return x.Id
	}
	return ""
}

func (x *IPAMRequest) GetIfName() string {
	if x != nil {
		return x.IfName
	}
	return ""
}

func (x *IPAMRequest) GetNamespace() string {
	if x != nil {
		return x.Namespace
	}
	return ""
}

func (x *IPAMRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *IPAMRequest) GetUid() string {
	if x != nil {
		return x.Uid
	}
	return ""
}

type IPAMResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Ip string `protobuf:"bytes,1,opt,name=ip,proto3" json:"ip,omitempty"`
}

func (x *IPAMResponse) Reset() {
	*x = IPAMResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ipam_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *IPAMResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*IPAMResponse) ProtoMessage() {}

func (x *IPAMResponse) ProtoReflect() protoreflect.Message {
	mi := &file_ipam_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use IPAMResponse.ProtoReflect.Descriptor instead.
func (*IPAMResponse) Descriptor() ([]byte, []int) {
	return file_ipam_proto_rawDescGZIP(), []int{1}
}

func (x *IPAMResponse) GetIp() string {
	if x != nil {
		return x.Ip
	}
	return ""
}

var File_ipam_proto protoreflect.FileDescriptor

var file_ipam_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x69, 0x70, 0x61, 0x6d, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x02, 0x76, 0x31,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x93,
	0x01, 0x0a, 0x0b, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x18,
	0x0a, 0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x07, 0x63, 0x6f, 0x6d, 0x6d, 0x61, 0x6e, 0x64, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x02, 0x69, 0x64, 0x12, 0x16, 0x0a, 0x06, 0x69, 0x66, 0x4e, 0x61,
	0x6d, 0x65, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x69, 0x66, 0x4e, 0x61, 0x6d, 0x65,
	0x12, 0x1c, 0x0a, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x18, 0x04, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x09, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x12, 0x12,
	0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x75, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x03, 0x75, 0x69, 0x64, 0x22, 0x1e, 0x0a, 0x0c, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x70, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x02, 0x69, 0x70, 0x32, 0x92, 0x01, 0x0a, 0x09, 0x69, 0x70, 0x53, 0x65, 0x72, 0x76, 0x69,
	0x63, 0x65, 0x12, 0x42, 0x0a, 0x08, 0x41, 0x6c, 0x6c, 0x6f, 0x63, 0x61, 0x74, 0x65, 0x12, 0x0f,
	0x2e, 0x76, 0x31, 0x2e, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x10, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73,
	0x65, 0x22, 0x13, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0d, 0x22, 0x08, 0x2f, 0x76, 0x31, 0x2f, 0x69,
	0x70, 0x61, 0x6d, 0x3a, 0x01, 0x2a, 0x12, 0x41, 0x0a, 0x07, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73,
	0x65, 0x12, 0x0f, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x10, 0x2e, 0x76, 0x31, 0x2e, 0x49, 0x50, 0x41, 0x4d, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x22, 0x13, 0x82, 0xd3, 0xe4, 0x93, 0x02, 0x0d, 0x2a, 0x08, 0x2f, 0x76,
	0x31, 0x2f, 0x69, 0x70, 0x61, 0x6d, 0x3a, 0x01, 0x2a, 0x42, 0x0c, 0x5a, 0x0a, 0x2e, 0x3b, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x5f, 0x76, 0x31, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ipam_proto_rawDescOnce sync.Once
	file_ipam_proto_rawDescData = file_ipam_proto_rawDesc
)

func file_ipam_proto_rawDescGZIP() []byte {
	file_ipam_proto_rawDescOnce.Do(func() {
		file_ipam_proto_rawDescData = protoimpl.X.CompressGZIP(file_ipam_proto_rawDescData)
	})
	return file_ipam_proto_rawDescData
}

var file_ipam_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_ipam_proto_goTypes = []interface{}{
	(*IPAMRequest)(nil),  // 0: v1.IPAMRequest
	(*IPAMResponse)(nil), // 1: v1.IPAMResponse
}
var file_ipam_proto_depIdxs = []int32{
	0, // 0: v1.ipService.Allocate:input_type -> v1.IPAMRequest
	0, // 1: v1.ipService.Release:input_type -> v1.IPAMRequest
	1, // 2: v1.ipService.Allocate:output_type -> v1.IPAMResponse
	1, // 3: v1.ipService.Release:output_type -> v1.IPAMResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ipam_proto_init() }
func file_ipam_proto_init() {
	if File_ipam_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ipam_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPAMRequest); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_ipam_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*IPAMResponse); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_ipam_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_ipam_proto_goTypes,
		DependencyIndexes: file_ipam_proto_depIdxs,
		MessageInfos:      file_ipam_proto_msgTypes,
	}.Build()
	File_ipam_proto = out.File
	file_ipam_proto_rawDesc = nil
	file_ipam_proto_goTypes = nil
	file_ipam_proto_depIdxs = nil
}
