package proxy

import (
	"github.com/BurntSushi/toml"
	"path"
	"strings"
)

type Config struct {
	// api
	HttpAddr string
	// store
	StoreAddr string
	// directory
	DirectoryAddr string
	// download domain
	Domain string
	// location prefix
	Prefix string
	// file
	MaxFileSize int
}

// NewConfig new a config.
func NewConfig(conf string) (c *Config, err error) {
	c = new(Config)
	if _, err = toml.DecodeFile(conf, c); err != nil {
		return
	}
	// bfs,/bfs,/bfs/ convert to /bfs/
	if c.Prefix != "" {
		c.Prefix = path.Join("/", c.Prefix) + "/"
		// http://domain/ covert to http://domain
		c.Domain = strings.TrimRight(c.Domain, "/")
	}
	return
}
