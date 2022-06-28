package errors

const (
	// hbase
	RetHBase = 30100
	// id
	RetIdNotAvailable = 30200
	// store
	RetStoreNotAvailable = 30300
	// zookeeper
	RetEtcdDataError = 30400
)

var (
	// hbase
	ErrHBase = Error(RetHBase)
	// id
	ErrIdNotAvailable = Error(RetIdNotAvailable)
	// store
	ErrStoreNotAvailable = Error(RetStoreNotAvailable)
	// etcd
	ErrEtcdDataError = Error(RetEtcdDataError)
)
