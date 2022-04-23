package store

import (
	"context"
	"kitten/pkg/errors"
	"kitten/pkg/stat"
	"kitten/pkg/store/conf"
	"kitten/pkg/store/needle"

	pb "kitten/api/store"
)

type Server struct {
	store *Store
	conf  *conf.Config
	info  *stat.Info
	pb.UnimplementedStoreServer
}

func NewStoreServer(store *Store, conf *conf.Config, info *stat.Info) *Server {
	return &Server{
		store: store,
		conf:  conf,
		info:  info,
	}
}

// GetFile implement grpc method, get file from Store.
func (s *Server) GetFile(ctx context.Context, req *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	if v := s.store.Volumes[req.Vid]; v != nil {
		if n, err := v.Read(req.Key, req.Cookie); err == nil {
			return &pb.GetFileResponse{Data: n.Data}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, errors.ErrVolumeNotExist
	}
}

// UploadFile implement grpc method, upload file to Store.
func (s *Server) UploadFile(ctx context.Context, req *pb.UploadFileRequest) (*pb.UploadFileResponse, error) {
	if len(req.Data) > s.conf.NeedleMaxSize {
		return nil, errors.ErrNeedleTooLarge
	}
	if v := s.store.Volumes[req.Vid]; v != nil {
		n := needle.NewWriter(req.Key, req.Cookie, int32(len(req.Data)))
		if err := n.ReadFromBytes(req.Data); err != nil {
			return nil, err
		}
		return &pb.UploadFileResponse{Message: "ok"}, nil
	} else {
		return nil, errors.ErrVolumeNotExist
	}
}

// DeleteFile implement grpc method, delete file from Store.
func (s *Server) DeleteFile(ctx context.Context, req *pb.DeleteFileRequest) (*pb.DeleteFileResponse, error) {
	if v := s.store.Volumes[req.Vid]; v != nil {
		if err := v.Delete(req.Key); err == nil {
			return &pb.DeleteFileResponse{Message: "ok"}, nil
		} else {
			return nil, err
		}
	} else {
		return nil, errors.ErrVolumeNotExist
	}
}
