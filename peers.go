package element

import (
	"github.com/gogo/protobuf/proto"
	"github.com/pkg/errors"
)

// Peers returns all known peers in the cluster
func (a *Agent) Peers() ([]*Peer, error) {
	peers := []*Peer{}
	members := a.members.Members()
	for _, node := range members {
		var s State
		if err := proto.Unmarshal(node.Meta, &s); err != nil {
			return nil, errors.Wrapf(err, "error unmarshalling proto for peer %s", node.Name)
		}
		peers = append(peers, s.Self)
	}
	return peers, nil
}

// Self returns the local peer information
func (a *Agent) Self() *Peer {
	return a.state.Self
}
