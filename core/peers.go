package core

import "github/mycache/pb"

// PeerPicker peeks a peer by key, usually implemented as
// a peers pool
type PeerPicker interface {
	Pick(key string) (peer Peer, ok bool)
}

// Peer is a cache node that has many groups
type Peer interface {
	// Fetch looks up key in group
	Fetch(in *pb.Request, out *pb.Response) error
}
