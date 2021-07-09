package core

// PeerPicker peeks a peer by key, usually implemented as
// a peers pool
type PeerPicker interface {
	Pick(key string) (peer Peer, ok bool)
}

type Peer interface {
	Get(group string, key string) ([]byte, error)
}
