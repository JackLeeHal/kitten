package main

import (
	"flag"
	"kitten/pkg/proxy"
	"log"
)

var (
	configFile string
)

func init() {
	flag.StringVar(&configFile, "c", "../proxy.toml", " set proxy config file path")
}

func main() {
	c, err := proxy.NewConfig(configFile)
	if err != nil {
		log.Fatalf("new config failed, err:%v", err)
	}

	server := proxy.NewServer(c)

	server.Start()
}
