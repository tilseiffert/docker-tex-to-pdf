// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.3.0
// - protoc             v4.23.4
// source: tex_to_pdf.proto

package protobuf

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

const (
	TexCompiler_CompileToPDF_FullMethodName = "/tex_to_pdf.TexCompiler/CompileToPDF"
)

// TexCompilerClient is the client API for TexCompiler service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type TexCompilerClient interface {
	CompileToPDF(ctx context.Context, in *CompileRequest, opts ...grpc.CallOption) (*CompileReply, error)
}

type texCompilerClient struct {
	cc grpc.ClientConnInterface
}

func NewTexCompilerClient(cc grpc.ClientConnInterface) TexCompilerClient {
	return &texCompilerClient{cc}
}

func (c *texCompilerClient) CompileToPDF(ctx context.Context, in *CompileRequest, opts ...grpc.CallOption) (*CompileReply, error) {
	out := new(CompileReply)
	err := c.cc.Invoke(ctx, TexCompiler_CompileToPDF_FullMethodName, in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// TexCompilerServer is the server API for TexCompiler service.
// All implementations must embed UnimplementedTexCompilerServer
// for forward compatibility
type TexCompilerServer interface {
	CompileToPDF(context.Context, *CompileRequest) (*CompileReply, error)
	mustEmbedUnimplementedTexCompilerServer()
}

// UnimplementedTexCompilerServer must be embedded to have forward compatible implementations.
type UnimplementedTexCompilerServer struct {
}

func (UnimplementedTexCompilerServer) CompileToPDF(context.Context, *CompileRequest) (*CompileReply, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CompileToPDF not implemented")
}
func (UnimplementedTexCompilerServer) mustEmbedUnimplementedTexCompilerServer() {}

// UnsafeTexCompilerServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to TexCompilerServer will
// result in compilation errors.
type UnsafeTexCompilerServer interface {
	mustEmbedUnimplementedTexCompilerServer()
}

func RegisterTexCompilerServer(s grpc.ServiceRegistrar, srv TexCompilerServer) {
	s.RegisterService(&TexCompiler_ServiceDesc, srv)
}

func _TexCompiler_CompileToPDF_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompileRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TexCompilerServer).CompileToPDF(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: TexCompiler_CompileToPDF_FullMethodName,
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TexCompilerServer).CompileToPDF(ctx, req.(*CompileRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// TexCompiler_ServiceDesc is the grpc.ServiceDesc for TexCompiler service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var TexCompiler_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tex_to_pdf.TexCompiler",
	HandlerType: (*TexCompilerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "CompileToPDF",
			Handler:    _TexCompiler_CompileToPDF_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "tex_to_pdf.proto",
}
