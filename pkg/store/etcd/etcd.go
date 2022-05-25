package etcd

import (
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kitten/pkg/log"
	"kitten/pkg/meta"
	"kitten/pkg/store/conf"
	"path"
	"strconv"
	"time"
)

type Client struct {
	path string
	conf *conf.Config
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

func (c *Client) SetStore(ctx context.Context, s *meta.Store) (err error) {
	var (
		saveData []byte
		//stat *stat.Stat
		os = new(meta.Store)
	)
	s.Id = c.conf.Etcd.ServerId
	s.Rack = c.conf.Etcd.Rack
	s.Status = meta.StoreStatusInit
	data, err := c.Get(ctx, c.path)
	if err != nil {
		log.Logger.Errorf("etcd.Get(\"%s\") error(%v)", c.path, err)
		return
	}
	if len(data.Kvs) > 0 {
		if err = json.Unmarshal(data.Kvs[0].Value, os); err != nil {
			log.Logger.Errorf("json.Unmarshal() error(%v)", err)
			return
		}
		log.Logger.Infof("\nold store meta: %s, \ncurrent store meta: %s", os, s)
		s.Status = os.Status
	}
	// meta.Status not modifify, may update by pitchfork
	if saveData, err = json.Marshal(s); err != nil {
		log.Logger.Errorf("json.Marshal() error(%v)", err)
		return
	}
	if _, err = c.Put(ctx, c.path, string(saveData)); err != nil {
		log.Logger.Errorf("zk.Set(\"%s\") error(%v)", c.path, err)
	}
	return
}

func (c *Client) Volumes(ctx context.Context) (lines []string, err error) {
	resp, err := c.Client.Get(ctx, c.path)
	for _, v := range resp.Kvs {
		lines = append(lines, string(v.Value))
	}

	return
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
