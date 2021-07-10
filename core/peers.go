package core

// PeerPicker peeks a peer by key, usually implemented as
// a peers pool
type PeerPicker interface {
	Pick(key string) (peer Peer, ok bool)
}

// Peer is a cache node that has many groups
type Peer interface {
	// Get looks up key in group
	Get(group string, key string) ([]byte, error)
}
