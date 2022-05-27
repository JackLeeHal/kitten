package store

import (
	"context"
	"fmt"
	"io/ioutil"
	"kitten/pkg/errors"
	"kitten/pkg/log"
	"kitten/pkg/meta"
	"kitten/pkg/store/conf"
	"kitten/pkg/store/etcd"
	myos "kitten/pkg/store/os"
	"kitten/pkg/store/volume"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Store get all volume metadata from an index file. index contains volume id,
// volume file path, the super block file index ends with ".idx" if the super
// block is /bfs/super_block_1, then the super block index file is
// /bfs/super_block_1.idx.
//
// volume index file format:
//  ---------------------------------
// | block_path,index_path,volume_id |
// | /bfs/block_1,/bfs/block_1.idx,1\n |
// | /bfs/block_2,/bfs/block_2.idx,2\n |
//  ---------------------------------
//
// store -> N volumes
//		 -> volume index -> volume info
//
// volume -> super block -> needle -> photo info
//        -> block index -> needle -> photo info without raw data

const (
	freeVolumePrefix = "_free_block_"
	volumeIndexExt   = ".idx"
	volumeFreeId     = -1
)

var (
	_compactSleep = time.Second * 10
)

// Store save volumes.
type Store struct {
	vf          *os.File
	fvf         *os.File
	FreeId      int32
	Volumes     map[int32]*volume.Volume // split volumes lock
	FreeVolumes []*volume.Volume
	etcd        *etcd.Client
	conf        *conf.Config
	flock       sync.Mutex // protect FreeId & saveIndex
	vlock       sync.Mutex // protect Volumes map
}

// NewStore
func NewStore(c *conf.Config) (*Store, error) {
	e, err := etcd.NewEtcd(c)
	if err != nil {
		return nil, err
	}
	s := &Store{
		etcd:    e,
		conf:    c,
		FreeId:  0,
		Volumes: make(map[int32]*volume.Volume),
	}
	if s.vf, err = os.OpenFile(c.Store.VolumeIndex, os.O_RDWR|os.O_CREATE, 0664); err != nil {
		log.Logger.Errorf("os.OpenFile(\"%s\") error(%v)", c.Store.VolumeIndex, err)
		s.Close()

		return nil, err
	}
	if s.fvf, err = os.OpenFile(c.Store.FreeVolumeIndex, os.O_RDWR|os.O_CREATE, 0664); err != nil {
		log.Logger.Errorf("os.OpenFile(\"%s\") error(%v)", c.Store.FreeVolumeIndex, err)
		s.Close()

		return nil, err
	}
	if err = s.init(); err != nil {
		s.Close()
		return nil, err
	}

	return s, nil
}

// init init the store.
func (s *Store) init() (err error) {
	if err = s.parseFreeVolumeIndex(); err == nil {
		err = s.parseVolumeIndex(context.Background())
	}
	return
}

func (s *Store) SetEtcd(ctx context.Context) error {
	if err := s.etcd.SetStore(ctx, &meta.Store{
		Stat:  s.conf.StatListen,
		Admin: s.conf.AdminListen,
		Api:   s.conf.ApiListen,
	}); err != nil {
		return err
	}
	if err := s.etcd.SetRoot(ctx); err != nil {
		return err
	}

	return nil
}

// parseFreeVolumeIndex parse free index from local.
func (s *Store) parseFreeVolumeIndex() (err error) {
	var (
		i     int
		id    int32
		bfile string
		ifile string
		v     *volume.Volume
		data  []byte
		ids   []int32
		lines []string
		bfs   []string
		ifs   []string
	)
	if data, err = ioutil.ReadAll(s.fvf); err != nil {
		log.Logger.Errorf("ioutil.ReadAll() error(%v)", err)
		return
	}
	lines = strings.Split(string(data), "\n")
	if _, ids, bfs, ifs, err = s.parseIndex(lines); err != nil {
		return
	}
	for i = 0; i < len(bfs); i++ {
		id, bfile, ifile = ids[i], bfs[i], ifs[i]
		if v, err = newVolume(id, bfile, ifile, s.conf); err != nil {
			return
		}
		v.Close()
		s.FreeVolumes = append(s.FreeVolumes, v)
		if id = s.fileFreeId(bfile); id > s.FreeId {
			s.FreeId = id
		}
	}
	log.Logger.Infof("current max free volume id: %d", s.FreeId)
	return
}

// parseVolumeIndex parse index from local config and etcd.
func (s *Store) parseVolumeIndex(ctx context.Context) (err error) {
	var (
		i          int
		ok         bool
		id         int32
		bfile      string
		ifile      string
		v          *volume.Volume
		data       []byte
		lids, zids []int32
		lines      []string
		lbfs, lifs []string
		zbfs, zifs []string
		lim        map[int32]struct{}
		zim        map[int32]struct{}
	)
	if data, err = ioutil.ReadAll(s.vf); err != nil {
		log.Logger.Errorf("ioutil.ReadAll() error(%v)", err)
		return
	}
	lines = strings.Split(string(data), "\n")
	if lim, lids, lbfs, lifs, err = s.parseIndex(lines); err != nil {
		return
	}
	if lines, err = s.etcd.Volumes(ctx); err != nil {
		return
	}
	if zim, zids, zbfs, zifs, err = s.parseIndex(lines); err != nil {
		return
	}
	// local index
	for i = 0; i < len(lbfs); i++ {
		id, bfile, ifile = lids[i], lbfs[i], lifs[i]
		if _, ok = s.Volumes[id]; ok {
			continue
		}
		if v, err = newVolume(id, bfile, ifile, s.conf); err != nil {
			return
		}
		s.Volumes[id] = v
		if _, ok = zim[id]; !ok {
			if err = s.etcd.AddVolume(ctx, id, v.Meta()); err != nil {
				return
			}
		} else {
			if err = s.etcd.SetVolume(ctx, id, v.Meta()); err != nil {
				return
			}
		}
	}
	// etcd index
	for i = 0; i < len(zbfs); i++ {
		id, bfile, ifile = zids[i], zbfs[i], zifs[i]
		if _, ok = s.Volumes[id]; ok {
			continue
		}
		// if not exists in local
		if _, ok = lim[id]; !ok {
			if v, err = newVolume(id, bfile, ifile, s.conf); err != nil {
				return
			}
			s.Volumes[id] = v
		}
	}
	err = s.saveVolumeIndex()
	return
}

// parseIndex parse volume info from a index file.
func (s *Store) parseIndex(lines []string) (im map[int32]struct{}, ids []int32, bfs, ifs []string, err error) {
	var (
		id    int64
		vid   int32
		line  string
		bfile string
		ifile string
		seps  []string
	)
	im = make(map[int32]struct{})
	for _, line = range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}
		if seps = strings.Split(line, ","); len(seps) != 3 {
			log.Logger.Errorf("volume index: \"%s\" format error", line)
			err = errors.ErrStoreVolumeIndex
			return
		}
		bfile = seps[0]
		ifile = seps[1]
		if id, err = strconv.ParseInt(seps[2], 10, 32); err != nil {
			log.Logger.Errorf("volume index: \"%s\" format error", line)
			return
		}
		vid = int32(id)
		ids = append(ids, vid)
		bfs = append(bfs, bfile)
		ifs = append(ifs, ifile)
		im[vid] = struct{}{}
		log.Logger.Infof("parse volume index, id: %d, block: %s, index: %s", id, bfile, ifile)
	}
	return
}

// saveFreeVolumeIndex save free volumes index info to disk.
func (s *Store) saveFreeVolumeIndex() (err error) {
	var (
		tn, n int
		v     *volume.Volume
	)
	if _, err = s.fvf.Seek(0, os.SEEK_SET); err != nil {
		log.Logger.Errorf("fvf.Seek() error(%v)", err)
		return
	}
	for _, v = range s.FreeVolumes {
		if n, err = s.fvf.WriteString(fmt.Sprintf("%s\n", string(v.Meta()))); err != nil {
			log.Logger.Errorf("fvf.WriteString() error(%v)", err)
			return
		}
		tn += n
	}
	if err = s.fvf.Sync(); err != nil {
		log.Logger.Errorf("fvf.saveFreeVolumeIndex Sync() error(%v)", err)
		return
	}
	if err = os.Truncate(s.conf.Store.FreeVolumeIndex, int64(tn)); err != nil {
		log.Logger.Errorf("os.Truncate() error(%v)", err)
	}
	return
}

// saveVolumeIndex save volumes index info to disk.
func (s *Store) saveVolumeIndex() (err error) {
	var (
		tn, n int
		v     *volume.Volume
	)
	if _, err = s.vf.Seek(0, os.SEEK_SET); err != nil {
		log.Logger.Errorf("vf.Seek() error(%v)", err)
		return
	}
	for _, v = range s.Volumes {
		if n, err = s.vf.WriteString(fmt.Sprintf("%s\n", string(v.Meta()))); err != nil {
			log.Logger.Errorf("vf.WriteString() error(%v)", err)
			return
		}
		tn += n
	}
	if err = s.vf.Sync(); err != nil {
		log.Logger.Errorf("vf.Sync() error(%v)", err)
		return
	}
	if err = os.Truncate(s.conf.Store.VolumeIndex, int64(tn)); err != nil {
		log.Logger.Errorf("os.Truncate() error(%v)", err)
	}
	return
}

// freeFile get volume block & index free file name.
func (s *Store) freeFile(id int32, bdir, idir string) (bfile, ifile string) {
	var file = fmt.Sprintf("%s%d", freeVolumePrefix, id)
	bfile = filepath.Join(bdir, file)
	file = fmt.Sprintf("%s%d%s", freeVolumePrefix, id, volumeIndexExt)
	ifile = filepath.Join(idir, file)
	return
}

// file get volume block & index file name.
func (s *Store) file(id int32, bdir, idir string, i int) (bfile, ifile string) {
	var file = fmt.Sprintf("%d_%d", id, i)
	bfile = filepath.Join(bdir, file)
	file = fmt.Sprintf("%d_%d%s", id, i, volumeIndexExt)
	ifile = filepath.Join(idir, file)
	return
}

// fileFreeId get a file free id.
func (s *Store) fileFreeId(file string) (id int32) {
	var (
		fid    int64
		fidStr = strings.Replace(filepath.Base(file), freeVolumePrefix, "", -1)
	)
	fid, _ = strconv.ParseInt(fidStr, 10, 32)
	id = int32(fid)
	return
}

// AddFreeVolume add free volumes.
func (s *Store) AddFreeVolume(n int, bdir, idir string) (sn int, err error) {
	var (
		i            int
		bfile, ifile string
		v            *volume.Volume
	)
	s.flock.Lock()
	for i = 0; i < n; i++ {
		s.FreeId++
		bfile, ifile = s.freeFile(s.FreeId, bdir, idir)
		if myos.Exist(bfile) || myos.Exist(ifile) {
			continue
		}
		if v, err = newVolume(volumeFreeId, bfile, ifile, s.conf); err != nil {
			// if no free space, delete the file
			os.Remove(bfile)
			os.Remove(ifile)
			break
		}
		v.Close()
		s.FreeVolumes = append(s.FreeVolumes, v)
		sn++
	}
	err = s.saveFreeVolumeIndex()
	s.flock.Unlock()
	return
}

// freeVolume get a free volume.
func (s *Store) freeVolume(id int32) (v *volume.Volume, err error) {
	var (
		i                                        int
		bfile, nbfile, ifile, nifile, bdir, idir string
	)
	s.flock.Lock()
	defer s.flock.Unlock()
	if len(s.FreeVolumes) == 0 {
		err = errors.ErrStoreNoFreeVolume
		return
	}
	v = s.FreeVolumes[0]
	s.FreeVolumes = s.FreeVolumes[1:]
	v.Id = id
	bfile, ifile = v.Block.File, v.Indexer.File
	bdir, idir = filepath.Dir(bfile), filepath.Dir(ifile)
	for {
		nbfile, nifile = s.file(id, bdir, idir, i)
		if !myos.Exist(nbfile) && !myos.Exist(nifile) {
			break
		}
		i++
	}
	log.Logger.Infof("rename block: %s to %s", bfile, nbfile)
	log.Logger.Infof("rename index: %s to %s", ifile, nifile)
	if err = os.Rename(ifile, nifile); err != nil {
		log.Logger.Errorf("os.Rename(\"%s\", \"%s\") error(%v)", ifile, nifile, err)
		v.Destroy()
		return
	}
	if err = os.Rename(bfile, nbfile); err != nil {
		log.Logger.Errorf("os.Rename(\"%s\", \"%s\") error(%v)", bfile, nbfile, err)
		v.Destroy()
		return
	}
	v.Block.File = nbfile
	v.Indexer.File = nifile
	if err = v.Open(); err != nil {
		v.Destroy()
		return
	}
	err = s.saveFreeVolumeIndex()
	return
}

// addVolume atomic add volume by copy-on-write.
func (s *Store) addVolume(id int32, nv *volume.Volume) {
	var (
		vid     int32
		v       *volume.Volume
		volumes = make(map[int32]*volume.Volume, len(s.Volumes)+1)
	)
	for vid, v = range s.Volumes {
		volumes[vid] = v
	}
	volumes[id] = nv
	// goroutine safe replace
	s.Volumes = volumes
}

// AddVolume add a new volume.
func (s *Store) AddVolume(ctx context.Context, id int32) (v *volume.Volume, err error) {
	var ov *volume.Volume
	// try check exists
	if ov = s.Volumes[id]; ov != nil {
		return nil, errors.ErrVolumeExist
	}
	// find a free volume
	if v, err = s.freeVolume(id); err != nil {
		return
	}
	s.vlock.Lock()
	if ov = s.Volumes[id]; ov == nil {
		s.addVolume(id, v)
		if err = s.saveVolumeIndex(); err == nil {
			err = s.etcd.AddVolume(ctx, id, v.Meta())
		}
		if err != nil {
			log.Logger.Errorf("add volume: %d error(%v), local index or etcd index may save failed", id, err)
		}
	} else {
		err = errors.ErrVolumeExist
	}
	s.vlock.Unlock()
	if err == errors.ErrVolumeExist {
		v.Destroy()
	}
	return
}

// delVolume atomic del volume by copy-on-write.
func (s *Store) delVolume(id int32) {
	var (
		vid     int32
		v       *volume.Volume
		volumes = make(map[int32]*volume.Volume, len(s.Volumes)-1)
	)
	for vid, v = range s.Volumes {
		volumes[vid] = v
	}
	delete(volumes, id)
	// goroutine safe replace
	s.Volumes = volumes
}

// DelVolume del the volume by volume id.
func (s *Store) DelVolume(ctx context.Context, id int32) (err error) {
	var v *volume.Volume
	s.vlock.Lock()
	if v = s.Volumes[id]; v != nil {
		if !v.Compact {
			s.delVolume(id)
			if err = s.saveVolumeIndex(); err == nil {
				err = s.etcd.DelVolume(ctx, id)
			}
			if err != nil {
				log.Logger.Errorf("del volume: %d error(%v), local index or etcd index may save failed", id, err)
			}
		} else {
			err = errors.ErrVolumeInCompact
		}
	} else {
		err = errors.ErrVolumeNotExist
	}
	s.vlock.Unlock()
	// if succced or not meta data saved error, close volume
	if err == nil || (err != errors.ErrVolumeInCompact &&
		err != errors.ErrVolumeNotExist) {
		v.Destroy()
	}
	return
}

// BulkVolume copy a super block from another store server add to this server.
func (s *Store) BulkVolume(ctx context.Context, id int32, bfile, ifile string) (err error) {
	var v, nv *volume.Volume
	// recovery new block
	if nv, err = newVolume(id, bfile, ifile, s.conf); err != nil {
		return
	}
	s.vlock.Lock()
	if v = s.Volumes[id]; v == nil {
		s.addVolume(id, nv)
		if err = s.saveVolumeIndex(); err == nil {
			err = s.etcd.AddVolume(ctx, id, nv.Meta())
		}
		if err != nil {
			log.Logger.Errorf("bulk volume: %d error(%v), local index or etcd index may save failed", id, err)
		}
	} else {
		err = errors.ErrVolumeExist
	}
	s.vlock.Unlock()
	return
}

// CompactVolume compact a super block to another file.
func (s *Store) CompactVolume(ctx context.Context, id int32) (err error) {
	var (
		v, nv      *volume.Volume
		bdir, idir string
	)
	// try check volume
	if v = s.Volumes[id]; v != nil {
		if v.Compact {
			return errors.ErrVolumeInCompact
		}
	} else {
		return errors.ErrVolumeExist
	}
	// find a free volume
	if nv, err = s.freeVolume(id); err != nil {
		return
	}
	log.Logger.Infof("start compact volume: (%d) %s to %s", id, v.Block.File, nv.Block.File)
	// no lock here, Compact is no side-effect
	if err = v.StartCompact(nv); err != nil {
		nv.Destroy()
		v.StopCompact(nil)
		return
	}
	s.vlock.Lock()
	if v = s.Volumes[id]; v != nil {
		log.Logger.Infof("stop compact volume: (%d) %s to %s", id, v.Block.File, nv.Block.File)
		if err = v.StopCompact(nv); err == nil {
			if err = s.saveVolumeIndex(); err == nil {
				err = s.etcd.SetVolume(ctx, id, v.Meta())
			}
			if err != nil {
				log.Logger.Errorf("compact volume: %d error(%v), local index or etcd index may save failed", id, err)
			}
		}
	} else {
		// never happen
		err = errors.ErrVolumeExist
		log.Logger.Errorf("compact volume: %d not exist(may bug)", id)
	}
	s.vlock.Unlock()
	// WARN if failed, nv is free volume, if succeed nv replace with v.
	// Sleep untill anyone had old volume variables all processed.
	time.Sleep(_compactSleep)
	nv.Destroy()
	if err == nil {
		bdir, idir = filepath.Dir(nv.Block.File), filepath.Dir(nv.Indexer.File)
		_, err = s.AddFreeVolume(1, bdir, idir)
	}
	return
}

// Close close the store.
// WARN the global variable store must first set nil and reject any other
// requests then safty close.
func (s *Store) Close() {
	log.Logger.Info("store close")
	var v *volume.Volume
	if s.vf != nil {
		s.vf.Close()
	}
	if s.fvf != nil {
		s.fvf.Close()
	}
	for _, v = range s.Volumes {
		log.Logger.Infof("volume[%d] close", v.Id)
		v.Close()
	}
	if s.etcd != nil {
		s.etcd.Close()
	}

	return
}

func newVolume(id int32, bfile, ifile string, c *conf.Config) (v *volume.Volume, err error) {
	if v, err = volume.NewVolume(id, bfile, ifile, c); err != nil {
		log.Logger.Errorf("newVolume(%d, %s, %s) error(%v)", id, bfile, ifile, err)
	}
	return
}
