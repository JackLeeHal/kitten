package proxy

import (
	dpb "kitten/api/directory"
	pb "kitten/api/store"
)

type Server struct {
	store     pb.StoreClient
	directory dpb.DirectoryClient
}

func NewServer(conf *Config) *Server {
	return nil
}
