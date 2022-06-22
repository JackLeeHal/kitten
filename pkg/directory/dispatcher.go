package directory

import (
	"kitten/pkg/errors"
	"kitten/pkg/log"
	"kitten/pkg/meta"
	"math/rand"
	"sync"
	"time"
)

// Dispatcher ,
// get raw data and processed into memory for http reqs
type Dispatcher struct {
	gids    []int            // for write eg:  gid:1;2   gids: [1,1,2,2,2,2,2]
	wrtVids map[string]int32 // choose most suitable written volume, always order by rest space.
	rand    *rand.Rand
	rlock   sync.Mutex
}

const (
	maxScore          = 1000
	nsToMs            = 1000000             // ns ->  us
	spaceBenchmark    = meta.MaxBlockOffset // 1 volume
	addDelayBenchmark = 100                 // 100ms   <100ms means no load, -Score==0
	baseAddDelay      = 100                 // 1s score:   -(1000/baseAddDelay)*addDelayBenchmark == -1000
	minFreeSpace      = 10 * 1024 * 1024    // 10M * PaddingSize every volume must have 80M left. PaddingSize:8
)

// NewDispatcher .
func NewDispatcher() (d *Dispatcher) {
	d = new(Dispatcher)
	d.rand = rand.New(rand.NewSource(time.Now().UnixNano()))
	return
}

// Update when zk updates
func (d *Dispatcher) Update(group map[int][]string,
	store map[string]*meta.Store, volume map[int32]*meta.VolumeState,
	storeVolume map[string][]int32) (err error) {
	var (
		gid                        int
		i                          int
		vid                        int32
		gids                       []int
		wrtVids                    map[string]int32
		sid                        string
		stores                     []string
		restSpace, minScore, score int
		totalAdd, totalAddDelay    uint64
		write, ok                  bool
		storeMeta                  *meta.Store
		volumeState                *meta.VolumeState
	)
	gids = []int{}
	wrtVids = map[string]int32{}
	for gid, stores = range group {
		write = true
		// check all stores can writeable by the group.
		for _, sid = range stores {
			if storeMeta, ok = store[sid]; !ok {
				log.Logger.Errorf("idStore cannot match store: %s", sid)
				break
			}
			if storeMeta == nil {
				log.Logger.Warnf("storeMeta is null, %s", sid)
				return
			}
			if !storeMeta.CanWrite() {
				write = false
				break
			}
		}
		if !write {
			continue
		}
		// calc score
		for _, sid = range stores {
			totalAdd, totalAddDelay, restSpace, minScore = 0, 0, 0, 0
			// get all volumes by the store.
			for _, vid = range storeVolume[sid] {
				volumeState = volume[vid]
				if volumeState == nil {
					log.Logger.Warnf("volumeState is nil, %d", vid)
					return
				}
				totalAdd = totalAdd + volumeState.TotalWriteProcessed
				restSpace = restSpace + int(volumeState.FreeSpace)
				totalAddDelay = totalAddDelay + volumeState.TotalWriteDelay
				// cacl most suitable written vid
				if volumeState.FreeSpace > minFreeSpace {
					if value, ok := wrtVids[sid]; !ok || vid < value {
						wrtVids[sid] = vid
					}
				}
			}
			score = d.calScore(int(totalAdd), int(totalAddDelay), restSpace)
			if score < minScore || minScore == 0 {
				minScore = score
			}
		}
		for i = 0; i < minScore; i++ {
			gids = append(gids, gid)
		}
	}
	d.gids = gids
	d.wrtVids = wrtVids
	return
}

// cal_score algorithm of calculating score
func (d *Dispatcher) calScore(totalAdd, totalAddDelay, restSpace int) (score int) {
	var (
		rsScore, adScore int
	)
	rsScore = (restSpace / int(spaceBenchmark))
	if rsScore > maxScore {
		rsScore = maxScore // more than 32T will be 32T and set score maxScore; less than 32G will be set 0 score;
	}
	if totalAdd != 0 {
		adScore = (((totalAddDelay / nsToMs) / totalAdd) / addDelayBenchmark) * baseAddDelay
		if adScore > maxScore {
			adScore = maxScore // more than 1s will be 1s and set score -maxScore; less than 100ms will be set -0 score;
		}
	}
	score = rsScore - adScore
	return
}

// VolumeID get a volume id.
func (d *Dispatcher) VolumeID(group map[int][]string, storeVolume map[string][]int32) (vid int32, err error) {
	var (
		stores []string
		gid    int
	)
	if len(d.gids) == 0 {
		err = errors.ErrStoreNotAvailable
		return
	}
	d.rlock.Lock()
	defer d.rlock.Unlock()
	gid = d.gids[d.rand.Intn(len(d.gids))]
	stores = group[gid]
	if len(stores) == 0 {
		err = errors.ErrZookeeperDataError
		return
	}
	vid = d.wrtVids[stores[0]]
	return
}
