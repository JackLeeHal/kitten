package directory

import (
	"context"
	pb "kitten/api/directory"
	"kitten/pkg/directory/conf"
	"kitten/pkg/errors"
	"kitten/pkg/log"
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

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	servers, ok := s.directory.volumeStore[req.Vid]
	if !ok {
		return nil, errors.ErrNeedleNotExist
	}
	stores := make([]string, 0, len(servers))
	for _, v := range servers {
		s, ok := s.directory.store[v]
		if !ok {
			log.Logger.Errorf("store cannot match: %s", v)
			continue
		}
		if !s.CanRead() {
			continue
		}
		stores = append(stores, s.Api)
	}

	return &pb.GetResponse{Stores: stores}, nil
}

func (s *Server) Upload(ctx context.Context, req *pb.UploadRequest) (*pb.UploadResponse, error) {
	vid, err := s.directory.dispatcher.VolumeID(s.directory.group, s.directory.storeVolume)
	if err != nil {
		log.Logger.Errorf("dispatcher.VolumeID error(%v)", err)

		return nil, errors.ErrStoreNotAvailable
	}
	servers := s.directory.volumeStore[vid]
	stores := make([]string, 0, len(servers))
	for _, v := range servers {
		s, ok := s.directory.store[v]
		if !ok {
			return nil, errors.ErrEtcdDataError
		}
		stores = append(stores, s.Api)
	}
	return nil, nil
}
