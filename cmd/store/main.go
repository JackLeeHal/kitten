package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	pb "kitten/api/store"
	"kitten/pkg/store"
	"kitten/pkg/store/conf"
	"log"
	"net"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "../store.toml", " set store config file path")
}

func main() {
	c, err := conf.NewConfig(configFile)
	if err != nil {
		log.Fatalf("new config failed, err:%v", err)
	}
	s, err := store.NewStore(c)
	if err != nil {
		log.Fatalf("new store failed, err:%v", err)
	}
	err = s.SetEtcd(context.Background())
	if err != nil {
		log.Fatalf("new store failed, err:%v", err)
	}
	lis, err := net.Listen("tcp", ":8086")
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterStoreServer(grpcServer, store.NewStoreServer(s, c, nil))

	grpcServer.Serve(lis)
}
