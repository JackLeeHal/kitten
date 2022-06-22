package etcd

import (
	"context"
	errorsx "github.com/pkg/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"kitten/pkg/directory/conf"
	"path"
	"time"
)

type Client struct {
	Watcher clientv3.Watcher
	*clientv3.Client
	conf *conf.Config
}

func NewClient(config *conf.Config) *Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   config.Etcd.Addrs,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	watcher := clientv3.NewWatcher(cli)

	return &Client{
		Client:  cli,
		Watcher: watcher,
	}
}

// WatchRacks get all racks and watch for changes.
func (c *Client) WatchRacks(ctx context.Context) ([]string, <-chan clientv3.WatchResponse, error) {
	ev := c.Watcher.Watch(ctx, c.conf.Etcd.StoreRoot)
	resp, err := c.Get(ctx, c.conf.Etcd.StoreRoot, clientv3.WithPrefix())
	if err != nil {
		return nil, nil, errorsx.Wrap(err, "Etcd.Get with prefix error")
	}
	var racks []string
	for _, v := range resp.Kvs {
		racks = append(racks, string(v.Value))
	}

	return racks, ev, nil
}

// Stores get all stores from etcd.
func (c *Client) Stores(ctx context.Context, rack string) ([]string, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.StoreRoot, rack), clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get stores error")
	}
	var stores []string
	for _, v := range resp.Kvs {
		stores = append(stores, string(v.Value))
	}

	return stores, nil
}

// Store get store metadata
func (c *Client) Store(ctx context.Context, rack, store string) ([]byte, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.StoreRoot, rack, store))
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get store metadata error")
	}

	return resp.Kvs[0].Value, nil
}

// StoreVolumes get volumes from a store
func (c *Client) StoreVolumes(ctx context.Context, rack, store string) ([]string, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.StoreRoot, rack, store), clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get stores error")
	}
	var volumes []string
	for _, v := range resp.Kvs {
		volumes = append(volumes, string(v.Value))
	}

	return volumes, nil
}

// Volumes get all Volumes from etcd.
func (c *Client) Volumes(ctx context.Context) ([]string, error) {
	resp, err := c.Get(ctx, c.conf.Etcd.VolumeRoot, clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get volumes with prefix error")
	}
	var volumes []string
	for _, v := range resp.Kvs {
		volumes = append(volumes, string(v.Value))
	}

	return volumes, nil
}

// Volume get volume metadata
func (c *Client) Volume(ctx context.Context, volume string) ([]byte, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.VolumeRoot, volume))
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get store metadata error")
	}

	return resp.Kvs[0].Value, nil
}

// VolumeStores get stores of volume
func (c *Client) VolumeStores(ctx context.Context, volume string) ([]string, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.VolumeRoot, volume), clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get VolumeStores with prefix error")
	}
	var stores []string
	for _, v := range resp.Kvs {
		stores = append(stores, string(v.Value))
	}

	return stores, nil
}

// Groups get all Groups from etcd.
func (c *Client) Groups(ctx context.Context) ([]string, error) {
	resp, err := c.Get(ctx, c.conf.Etcd.GroupRoot, clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get volumes with prefix error")
	}
	var groups []string
	for _, v := range resp.Kvs {
		groups = append(groups, string(v.Value))
	}

	return groups, nil
}

// GroupStores get stores of group
func (c *Client) GroupStores(ctx context.Context, group string) ([]string, error) {
	resp, err := c.Get(ctx, path.Join(c.conf.Etcd.GroupRoot, group), clientv3.WithPrefix())
	if err != nil {
		return nil, errorsx.Wrap(err, "etcd.Get VolumeStores with prefix error")
	}
	var stores []string
	for _, v := range resp.Kvs {
		stores = append(stores, string(v.Value))
	}

	return stores, nil
}
