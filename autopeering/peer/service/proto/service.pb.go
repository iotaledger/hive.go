// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.23.0
// 	protoc        v3.11.4
// source: autopeering/peer/service/proto/service.proto

package proto

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// Mapping between a service ID and its tuple network_address
// e.g., map[autopeering:&{tcp, 198.51.100.1:80}].
type ServiceMap struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Map map[string]*NetworkAddress `protobuf:"bytes,1,rep,name=map,proto3" json:"map,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *ServiceMap) Reset() {
	*x = ServiceMap{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autopeering_peer_service_proto_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ServiceMap) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ServiceMap) ProtoMessage() {}

func (x *ServiceMap) ProtoReflect() protoreflect.Message {
	mi := &file_autopeering_peer_service_proto_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}

		return ms
	}

	return mi.MessageOf(x)
}

// Deprecated: Use ServiceMap.ProtoReflect.Descriptor instead.
func (*ServiceMap) Descriptor() ([]byte, []int) {
	return file_autopeering_peer_service_proto_service_proto_rawDescGZIP(), []int{0}
}

func (x *ServiceMap) GetMap() map[string]*NetworkAddress {
	if x != nil {
		return x.Map
	}

	return nil
}

// The service type (e.g., tcp, upd) and the address (e.g., 198.51.100.1:80).
type NetworkAddress struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Network string `protobuf:"bytes,1,opt,name=network,proto3" json:"network,omitempty"`
	Port    uint32 `protobuf:"varint,2,opt,name=port,proto3" json:"port,omitempty"`
}

func (x *NetworkAddress) Reset() {
	*x = NetworkAddress{}
	if protoimpl.UnsafeEnabled {
		mi := &file_autopeering_peer_service_proto_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NetworkAddress) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NetworkAddress) ProtoMessage() {}

func (x *NetworkAddress) ProtoReflect() protoreflect.Message {
	mi := &file_autopeering_peer_service_proto_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}

		return ms
	}

	return mi.MessageOf(x)
}

// Deprecated: Use NetworkAddress.ProtoReflect.Descriptor instead.
func (*NetworkAddress) Descriptor() ([]byte, []int) {
	return file_autopeering_peer_service_proto_service_proto_rawDescGZIP(), []int{1}
}

func (x *NetworkAddress) GetNetwork() string {
	if x != nil {
		return x.Network
	}

	return ""
}

func (x *NetworkAddress) GetPort() uint32 {
	if x != nil {
		return x.Port
	}

	return 0
}

var File_autopeering_peer_service_proto_service_proto protoreflect.FileDescriptor

var file_autopeering_peer_service_proto_service_proto_rawDesc = []byte{
	0x0a, 0x2c, 0x61, 0x75, 0x74, 0x6f, 0x70, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70, 0x65,
	0x65, 0x72, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x05,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x89, 0x01, 0x0a, 0x0a, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x4d, 0x61, 0x70, 0x12, 0x2c, 0x0a, 0x03, 0x6d, 0x61, 0x70, 0x18, 0x01, 0x20, 0x03, 0x28,
	0x0b, 0x32, 0x1a, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x53, 0x65, 0x72, 0x76, 0x69, 0x63,
	0x65, 0x4d, 0x61, 0x70, 0x2e, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x03, 0x6d,
	0x61, 0x70, 0x1a, 0x4d, 0x0a, 0x08, 0x4d, 0x61, 0x70, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10,
	0x0a, 0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79,
	0x12, 0x2b, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x15, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x41,
	0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38,
	0x01, 0x22, 0x3e, 0x0a, 0x0e, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x41, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x12, 0x12, 0x0a,
	0x04, 0x70, 0x6f, 0x72, 0x74, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x04, 0x70, 0x6f, 0x72,
	0x74, 0x42, 0x3e, 0x5a, 0x3c, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x69, 0x6f, 0x74, 0x61, 0x6c, 0x65, 0x64, 0x67, 0x65, 0x72, 0x2f, 0x68, 0x69, 0x76, 0x65, 0x2e,
	0x67, 0x6f, 0x2f, 0x61, 0x75, 0x74, 0x6f, 0x70, 0x65, 0x65, 0x72, 0x69, 0x6e, 0x67, 0x2f, 0x70,
	0x65, 0x65, 0x72, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_autopeering_peer_service_proto_service_proto_rawDescOnce sync.Once
	file_autopeering_peer_service_proto_service_proto_rawDescData = file_autopeering_peer_service_proto_service_proto_rawDesc
)

func file_autopeering_peer_service_proto_service_proto_rawDescGZIP() []byte {
	file_autopeering_peer_service_proto_service_proto_rawDescOnce.Do(func() {
		file_autopeering_peer_service_proto_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_autopeering_peer_service_proto_service_proto_rawDescData)
	})

	return file_autopeering_peer_service_proto_service_proto_rawDescData
}

var file_autopeering_peer_service_proto_service_proto_msgTypes = make([]protoimpl.MessageInfo, 3)
var file_autopeering_peer_service_proto_service_proto_goTypes = []interface{}{
	(*ServiceMap)(nil),     // 0: proto.ServiceMap
	(*NetworkAddress)(nil), // 1: proto.NetworkAddress
	nil,                    // 2: proto.ServiceMap.MapEntry
}
var file_autopeering_peer_service_proto_service_proto_depIdxs = []int32{
	2, // 0: proto.ServiceMap.map:type_name -> proto.ServiceMap.MapEntry
	1, // 1: proto.ServiceMap.MapEntry.value:type_name -> proto.NetworkAddress
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	2, // [2:2] is the sub-list for extension type_name
	2, // [2:2] is the sub-list for extension extendee
	0, // [0:2] is the sub-list for field type_name
}

func init() { file_autopeering_peer_service_proto_service_proto_init() }
func file_autopeering_peer_service_proto_service_proto_init() {
	if File_autopeering_peer_service_proto_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_autopeering_peer_service_proto_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ServiceMap); i {
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
		file_autopeering_peer_service_proto_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NetworkAddress); i {
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
			RawDescriptor: file_autopeering_peer_service_proto_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   3,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_autopeering_peer_service_proto_service_proto_goTypes,
		DependencyIndexes: file_autopeering_peer_service_proto_service_proto_depIdxs,
		MessageInfos:      file_autopeering_peer_service_proto_service_proto_msgTypes,
	}.Build()
	File_autopeering_peer_service_proto_service_proto = out.File
	file_autopeering_peer_service_proto_service_proto_rawDesc = nil
	file_autopeering_peer_service_proto_service_proto_goTypes = nil
	file_autopeering_peer_service_proto_service_proto_depIdxs = nil
}
