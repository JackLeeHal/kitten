package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"path"
	"strconv"
	"time"
)

type Client struct {
	path string
	*clientv3.Client
}

func NewClient(endpoints []string, path string) *Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}

	return &Client{
		Client: cli,
		path:   path,
	}
}

func (c *Client) AddVolume(ctx context.Context, id int32, data []byte) error {
	_, err := c.Client.Put(ctx, c.volumePath(id), string(data))

	return err
}

func (c *Client) Close() {
	c.Client.Close()
}

func (c *Client) volumePath(id int32) string {
	return path.Join(c.path, strconv.Itoa(int(id)))
}
