package directory

import (
	"kitten/pkg/meta"
	"kitten/pkg/store/conf"
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

	config *conf.Config
}
