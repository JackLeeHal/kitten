package directory

import (
	"context"
	pb "kitten/api/directory"
	"kitten/pkg/directory/conf"
)

type Server struct {
	directory *Directory
	conf      *conf.Config
	pb.UnimplementedDirectoryServer
}

func NewDirectoryServer(d *Directory, conf *conf.Config) *Server {
	return &Server{
		directory: d,
		conf:      conf,
	}
}

func (s *Server) Get(context.Context, *pb.GetRequest) (*pb.GetResponse, error) {
	return nil, nil
}

func (s *Server) Upload(context.Context, *pb.UploadRequest) (*pb.UploadResponse, error) {
	return nil, nil
}
