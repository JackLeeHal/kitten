package directory

import (
	"context"
	"encoding/json"
	errorsx "github.com/pkg/errors"
	etcd "go.etcd.io/etcd/client/v3"
	"kitten/pkg/directory/conf"
	myetcd "kitten/pkg/directory/etcd"
	"kitten/pkg/log"
	"kitten/pkg/meta"
	"strconv"
	"time"
)

const (
	retrySleep = time.Second * 1
)

type Directory struct {
	// STORE
	store       map[string]*meta.Store // store_server_id:store_info
	storeVolume map[string][]int32     // store_server_id:volume_ids

	// GROUP
	storeGroup map[string]int   // store_server_id:group
	group      map[int][]string // group_id:store_servers

	// VOLUME
	volume      map[int32]*meta.VolumeState // volume_id:volume_state
	volumeStore map[int32][]string          // volume_id:store_server_id

	dispatcher *Dispatcher // dispatch for write or read reqs
	config     *conf.Config
	etcd       *myetcd.Client
}

func NewDirectory(config *conf.Config) (*Directory, error) {
	e := myetcd.NewClient(config)
	d := &Directory{
		etcd:       e,
		dispatcher: NewDispatcher(),
	}

	go d.SyncEtcd()

	return d, nil
}

// SyncEtcd Sync data from Etcd to memory
func (d *Directory) SyncEtcd() {
	ctx := context.TODO()
	for {
		evs, err := d.syncStores(ctx)
		if err != nil {
			log.Logger.Errorf("syncStores() called error(%v)", err)
			time.Sleep(retrySleep)
			continue
		}
		if err = d.syncGroups(ctx); err != nil {
			log.Logger.Errorf("syncGroups() called error(%v)", err)
			time.Sleep(retrySleep)
			continue
		}
		if err = d.syncVolumes(ctx); err != nil {
			log.Logger.Errorf("syncVolumes() called error(%v)", err)
			time.Sleep(retrySleep)
			continue
		}
		if err = d.dispatcher.Update(d.group, d.store, d.volume, d.storeVolume); err != nil {
			log.Logger.Errorf("Update() called error(%v)", err)
			time.Sleep(retrySleep)
			continue
		}
		select {
		case <-evs:
			log.Logger.Infof("stores status change or new store")
			break
		case <-time.After(d.config.Etcd.PullInterval.Duration):
			log.Logger.Infof("pull from zk")
			break
		}
	}
}

// syncStores get all the store nodes and set a watcher
func (d *Directory) syncStores(ctx context.Context) (<-chan etcd.WatchResponse, error) {
	racks, evs, err := d.etcd.WatchRacks(ctx)
	if err != nil {
		return nil, errorsx.WithMessage(err, "WatchRacks with error")
	}
	store := make(map[string]*meta.Store)
	storeVolume := make(map[string][]int32)
	for _, rack := range racks {
		// get all stores
		stores, err := d.etcd.Stores(ctx, rack)
		if err != nil {
			return nil, err
		}
		// get sore metadata
		for _, s := range stores {
			data, err := d.etcd.Store(ctx, rack, s)
			if err != nil {
				return nil, err
			}
			storeMeta := new(meta.Store)
			if err = json.Unmarshal(data, storeMeta); err != nil {
				return nil, errorsx.Wrap(err, "store metadata unmarshal error")
			}
			// get all volumes from store
			volumes, err := d.etcd.StoreVolumes(ctx, rack, s)
			if err != nil {
				return nil, err
			}
			storeVolume[storeMeta.Id] = []int32{}
			storeVolume[storeMeta.Id] = []int32{}
			for _, volume := range volumes {
				vid, err := strconv.Atoi(volume)
				if err != nil {
					log.Logger.Errorf("wrong volume:%s", volume)
					continue
				}
				storeVolume[storeMeta.Id] = append(storeVolume[storeMeta.Id], int32(vid))
			}
			store[storeMeta.Id] = storeMeta
		}
	}
	d.store = store
	d.storeVolume = storeVolume

	return evs, nil
}

func (d *Directory) syncGroups(ctx context.Context) error {
	groups, err := d.etcd.Groups(ctx)
	if err != nil {
		return err
	}
	group := make(map[int][]string)
	storeGroup := make(map[string]int)
	for _, s := range groups {
		// get all stores by the group
		stores, err := d.etcd.GroupStores(ctx, s)
		if err != nil {
			return err
		}
		gid, err := strconv.Atoi(s)
		if err != nil {
			log.Logger.Errorf("wrong group:%s", s)
			continue
		}
		group[gid] = stores
		for _, s = range stores {
			storeGroup[s] = gid
		}
	}
	d.group = group
	d.storeGroup = storeGroup

	return nil
}

func (d *Directory) syncVolumes(ctx context.Context) error {
	// get all volumes
	volumes, err := d.etcd.Volumes(ctx)
	if err != nil {
		return err
	}
	volume := make(map[int32]*meta.VolumeState)
	volumeStore := make(map[int32][]string)
	for _, s := range volumes {
		// get the volume
		data, err := d.etcd.Volume(ctx, s)
		if err != nil {
			return err
		}
		volumeState := new(meta.VolumeState)
		if err = json.Unmarshal(data, volumeState); err != nil {
			return errorsx.Wrap(err, "volume metadata unmarshal error")
		}
		vid, err := strconv.Atoi(s)
		if err != nil {
			log.Logger.Errorf("wrong volume:%s", s)
			continue
		}
		volume[int32(vid)] = volumeState
		// get the stores by the volume
		stores, err := d.etcd.VolumeStores(ctx, s)
		if err != nil {
			return err
		}
		volumeStore[int32(vid)] = stores
	}
	d.volume = volume
	d.volumeStore = volumeStore

	return nil
}
