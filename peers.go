package mycache

// PeerPicker peeks a peer by key
type PeerPicker interface {
	Pick(key string) (peer Peer, ok bool)
}

type Peer interface {
	Get(group string, key string) ([]byte, error)
}
