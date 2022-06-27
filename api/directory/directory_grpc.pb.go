// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.19.4
// source: directory/directory.proto

package directory

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// DirectoryClient is the client API for Directory service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type DirectoryClient interface {
	// Get store servers
	Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error)
	// Upload get vid and store servers for upload
	Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadResponse, error)
}

type directoryClient struct {
	cc grpc.ClientConnInterface
}

func NewDirectoryClient(cc grpc.ClientConnInterface) DirectoryClient {
	return &directoryClient{cc}
}

func (c *directoryClient) Get(ctx context.Context, in *GetRequest, opts ...grpc.CallOption) (*GetResponse, error) {
	out := new(GetResponse)
	err := c.cc.Invoke(ctx, "/directory.Directory/Get", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *directoryClient) Upload(ctx context.Context, in *UploadRequest, opts ...grpc.CallOption) (*UploadResponse, error) {
	out := new(UploadResponse)
	err := c.cc.Invoke(ctx, "/directory.Directory/Upload", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// DirectoryServer is the server API for Directory service.
// All implementations must embed UnimplementedDirectoryServer
// for forward compatibility
type DirectoryServer interface {
	// Get store servers
	Get(context.Context, *GetRequest) (*GetResponse, error)
	// Upload get vid and store servers for upload
	Upload(context.Context, *UploadRequest) (*UploadResponse, error)
	mustEmbedUnimplementedDirectoryServer()
}

// UnimplementedDirectoryServer must be embedded to have forward compatible implementations.
type UnimplementedDirectoryServer struct {
}

func (UnimplementedDirectoryServer) Get(context.Context, *GetRequest) (*GetResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Get not implemented")
}
func (UnimplementedDirectoryServer) Upload(context.Context, *UploadRequest) (*UploadResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Upload not implemented")
}
func (UnimplementedDirectoryServer) mustEmbedUnimplementedDirectoryServer() {}

// UnsafeDirectoryServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to DirectoryServer will
// result in compilation errors.
type UnsafeDirectoryServer interface {
	mustEmbedUnimplementedDirectoryServer()
}

func RegisterDirectoryServer(s grpc.ServiceRegistrar, srv DirectoryServer) {
	s.RegisterService(&Directory_ServiceDesc, srv)
}

func _Directory_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(GetRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DirectoryServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/directory.Directory/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DirectoryServer).Get(ctx, req.(*GetRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func _Directory_Upload_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(UploadRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(DirectoryServer).Upload(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/directory.Directory/Upload",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(DirectoryServer).Upload(ctx, req.(*UploadRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// Directory_ServiceDesc is the grpc.ServiceDesc for Directory service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Directory_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "directory.Directory",
	HandlerType: (*DirectoryServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Directory_Get_Handler,
		},
		{
			MethodName: "Upload",
			Handler:    _Directory_Upload_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "directory/directory.proto",
}
