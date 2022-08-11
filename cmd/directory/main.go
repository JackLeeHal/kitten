package main

import (
	"flag"
	"google.golang.org/grpc"
	pb "kitten/api/directory"
	"kitten/pkg/directory"
	"kitten/pkg/directory/conf"
	"log"
	"net"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "../directory.toml", " set directory config file path")
}

func main() {
	c, err := conf.NewConfig(configFile)
	if err != nil {
		log.Fatalf("new config failed, err:%v", err)
	}
	d, err := directory.NewDirectory(c)
	if err != nil {
		log.Fatalf("new store failed, err:%v", err)
	}

	lis, err := net.Listen("tcp", ":8086")
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterDirectoryServer(grpcServer, directory.NewDirectoryServer(d, c))

	grpcServer.Serve(lis)
}
