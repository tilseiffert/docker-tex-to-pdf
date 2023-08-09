package server

import (
	"context"
	"fmt"
	"net"
	"strconv"

	pb "github.com/tilseiffert/docker-tex-to-pdf/internal/protobuf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	StandardPort = 50051
)

// server is used to implement the TexCompilerServer interface
type server struct {
	pb.UnimplementedTexCompilerServer
}

// CompileToPDF is the implementation of the gRPC method CompileToPDF
func (s *server) CompileToPDF(ctx context.Context, req *pb.CompileRequest) (*pb.CompileReply, error) {
	// Hier wird Ihre Logik zum Kompilieren der TeX-Datei in PDF implementiert
	// ...

	// Ein einfaches Beispiel:
	pdfContent := []byte("This would be the binary content of the PDF.")
	logMessage := "Compiled successfully."

	return &pb.CompileReply{PdfContent: pdfContent, Log: logMessage}, nil
}

// Main starts the gRPC server on the given port
// If port is 0, the standard port 50051 is used
func Main(port int) (*grpc.Server, error) {

	if port == 0 {
		port = StandardPort
	}

	// Create a TCP listener on the given port
	lis, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	if err != nil {
		return nil, fmt.Errorf("failed create listener: %w", err)
	}

	// Create a new gRPC server
	s := grpc.NewServer()

	// Register the TexCompilerServer with the gRPC server
	pb.RegisterTexCompilerServer(s, &server{})

	// Register reflection service on gRPC server.
	reflection.Register(s)

	// Start the gRPC server
	if err := s.Serve(lis); err != nil {
		return nil, fmt.Errorf("failed to start grpc-server: %w", err)
	}

	return s, nil
}
