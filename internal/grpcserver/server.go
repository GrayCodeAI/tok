// Package grpcserver provides gRPC compression service.
package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/GrayCodeAI/tokman/internal/filter"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CompressionService implements the gRPC compression service.
type CompressionService struct {
	UnimplementedCompressionServiceServer
}

// UnimplementedCompressionServiceServer must be embedded for forward compatibility.
type UnimplementedCompressionServiceServer struct{}

func (UnimplementedCompressionServiceServer) Compress(context.Context, *CompressRequest) (*CompressResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Compress not implemented")
}

func (UnimplementedCompressionServiceServer) StreamCompress(CompressionService_StreamCompressServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamCompress not implemented")
}

func (UnimplementedCompressionServiceServer) GetStats(context.Context, *StatsRequest) (*StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method GetStats not implemented")
}

// CompressRequest represents a compression request.
type CompressRequest struct {
	Content string `protobuf:"bytes,1,opt,name=content,proto3" json:"content,omitempty"`
	Mode    string `protobuf:"bytes,2,opt,name=mode,proto3" json:"mode,omitempty"`
	Budget  int32  `protobuf:"varint,3,opt,name=budget,proto3" json:"budget,omitempty"`
}

// CompressResponse represents a compression response.
type CompressResponse struct {
	Compressed       string  `protobuf:"bytes,1,opt,name=compressed,proto3" json:"compressed,omitempty"`
	OriginalTokens   int32   `protobuf:"varint,2,opt,name=original_tokens,json=originalTokens,proto3" json:"original_tokens,omitempty"`
	CompressedTokens int32   `protobuf:"varint,3,opt,name=compressed_tokens,json=compressedTokens,proto3" json:"compressed_tokens,omitempty"`
	SavedTokens      int32   `protobuf:"varint,4,opt,name=saved_tokens,json=savedTokens,proto3" json:"saved_tokens,omitempty"`
	ReductionPercent float64 `protobuf:"fixed64,5,opt,name=reduction_percent,json=reductionPercent,proto3" json:"reduction_percent,omitempty"`
}

// StatsRequest represents a stats request.
type StatsRequest struct{}

// StatsResponse represents stats response.
type StatsResponse struct {
	TotalRequests    int64 `protobuf:"varint,1,opt,name=total_requests,json=totalRequests,proto3" json:"total_requests,omitempty"`
	TotalTokensSaved int64 `protobuf:"varint,2,opt,name=total_tokens_saved,json=totalTokensSaved,proto3" json:"total_tokens_saved,omitempty"`
}

// CompressionService_StreamCompressServer is the server API for streaming.
type CompressionService_StreamCompressServer interface {
	Recv() (*CompressRequest, error)
	Send(*CompressResponse) error
	grpc.ServerStream
}

// Compress handles single compression requests.
func (s *CompressionService) Compress(ctx context.Context, req *CompressRequest) (*CompressResponse, error) {
	if req.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}

	mode := filter.ModeMinimal
	if req.Mode == "aggressive" {
		mode = filter.ModeAggressive
	}

	engine := filter.NewEngine(mode)
	compressed, saved := engine.Process(req.Content)

	originalTokens := filter.EstimateTokens(req.Content)
	compressedTokens := filter.EstimateTokens(compressed)

	return &CompressResponse{
		Compressed:       compressed,
		OriginalTokens:   int32(originalTokens),
		CompressedTokens: int32(compressedTokens),
		SavedTokens:      int32(saved),
		ReductionPercent: float64(saved) / float64(originalTokens) * 100,
	}, nil
}

// StreamCompress handles streaming compression.
func (s *CompressionService) StreamCompress(stream CompressionService_StreamCompressServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			return err
		}

		resp, err := s.Compress(stream.Context(), req)
		if err != nil {
			return err
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

// GetStats returns service statistics.
func (s *CompressionService) GetStats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	return &StatsResponse{
		TotalRequests:    0, // Would track in production
		TotalTokensSaved: 0,
	}, nil
}

// Server provides gRPC server functionality.
type Server struct {
	grpcServer *grpc.Server
	listener   net.Listener
	port       int
}

// NewServer creates a new gRPC server.
func NewServer(port int) *Server {
	return &Server{port: port}
}

// Start starts the gRPC server.
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.listener = lis
	s.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024),
		grpc.MaxSendMsgSize(10*1024*1024),
	)

	RegisterCompressionServiceServer(s.grpcServer, &CompressionService{})

	go func() {
		if err := s.grpcServer.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	return nil
}

// Stop stops the gRPC server.
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}

// Port returns the actual port.
func (s *Server) Port() int {
	if s.listener != nil {
		return s.listener.Addr().(*net.TCPAddr).Port
	}
	return s.port
}

// RegisterCompressionServiceServer registers the service.
func RegisterCompressionServiceServer(s *grpc.Server, srv *CompressionService) {
	s.RegisterService(&CompressionService_ServiceDesc, srv)
}

// CompressionService_ServiceDesc is the grpc.ServiceDesc.
var CompressionService_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "tokman.CompressionService",
	HandlerType: (*CompressionService)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Compress",
			Handler:    CompressionService_Compress_Handler,
		},
		{
			MethodName: "GetStats",
			Handler:    CompressionService_GetStats_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamCompress",
			Handler:       CompressionService_StreamCompress_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "compression.proto",
}

// Handler implementations.
func CompressionService_Compress_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(CompressRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompressionServiceServer).Compress(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tokman.CompressionService/Compress",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompressionServiceServer).Compress(ctx, req.(*CompressRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func CompressionService_GetStats_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(CompressionServiceServer).GetStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/tokman.CompressionService/GetStats",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(CompressionServiceServer).GetStats(ctx, req.(*StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func CompressionService_StreamCompress_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(CompressionServiceServer).StreamCompress(&compressionServiceStreamCompressServer{stream})
}

type compressionServiceStreamCompressServer struct {
	grpc.ServerStream
}

func (x *compressionServiceStreamCompressServer) Recv() (*CompressRequest, error) {
	m := new(CompressRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (x *compressionServiceStreamCompressServer) Send(m *CompressResponse) error {
	return x.ServerStream.SendMsg(m)
}

// CompressionServiceServer is the server API.
type CompressionServiceServer interface {
	Compress(context.Context, *CompressRequest) (*CompressResponse, error)
	StreamCompress(CompressionService_StreamCompressServer) error
	GetStats(context.Context, *StatsRequest) (*StatsResponse, error)
}
