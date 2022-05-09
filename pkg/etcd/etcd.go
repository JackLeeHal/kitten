package etcd

import (
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kitten/pkg/log"
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

func (c *Client) SetVolume(ctx context.Context, id int32, data []byte) error {
	vPath := c.volumePath(id)
	if _, err := c.Get(ctx, vPath); err != nil {
		log.Logger.Errorf("etcd.Get(\"%s\") error(%v)", vPath, err)

		return err
	}
	_, err := c.Client.Put(ctx, vPath, string(data))
	if err != nil {
		log.Logger.Errorf("etcd.Put(\"%s\", \"%s\") error(%v)", vPath, string(data), err)

		return err
	}

	return nil
}

func (c *Client) DelVolume(ctx context.Context, id int32) error {
	vPath := c.volumePath(id)
	if _, err := c.Get(ctx, vPath); err != nil {
		log.Logger.Errorf("etcd.Get(\"%s\") error(%v)", vPath, err)

		return err
	}
	_, err := c.Client.Delete(ctx, vPath)
	if err != nil {
		log.Logger.Errorf("etcd.Delete(\"%s\") error(%v)", vPath, err)

		return err
	}

	return nil
}

func (c *Client) Close() {
	c.Client.Close()
}

func (c *Client) volumePath(id int32) string {
	return path.Join(c.path, strconv.Itoa(int(id)))
}
