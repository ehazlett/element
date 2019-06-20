package element

import (
	"time"

	"github.com/containerd/typeurl"
	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/sirupsen/logrus"
)

// NodeMeta returns local node meta information
func (a *Agent) NodeMeta(limit int) []byte {
	data, err := proto.Marshal(a.metadata)
	if err != nil {
		logrus.Errorf("error serializing node meta: %s", err)
	}
	return data
}

// NotifyMsg is used for handling cluster messages
func (a *Agent) NotifyMsg(buf []byte) {
	// this can be used to receive messages sent (i.e. SendReliable)
	t := &types.Any{}
	if err := t.Unmarshal(buf); err != nil {
		logrus.Errorf("error unmarshalling proto message: %s")
		return
	}
	v, err := typeurl.UnmarshalAny(t)
	if err != nil {
		logrus.Errorf("error unmarshalling from any: %s", err)
		return
	}
	h, ok := a.messageHandlers[t.TypeUrl]
	if !ok {
		logrus.Warnf("no message handler for type %s", t.TypeUrl)
		return
	}
	if err := h(v); err != nil {
		logrus.Errorf("error from message handler for type %s: %s", t.TypeUrl, err)
		return
	}
}

// GetBroadcasts is called when user messages can be broadcast
func (a *Agent) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

// LocalState is the local cluster agent state
func (a *Agent) LocalState(join bool) []byte {
	data, err := proto.Marshal(a.metadata)
	if err != nil {
		logrus.Errorf("error serializing local state: %s", err)
	}
	return data
}

// MergeRemoteState is used to store remote peer information
func (a *Agent) MergeRemoteState(buf []byte, join bool) {
	var meta Metadata
	if err := proto.Unmarshal(buf, &meta); err != nil {
		logrus.Errorf("error parsing remote agent meta: %s", err)
		return
	}
	logrus.Debugf("merge remote state: %+v", meta)
	a.metadata.Updated = time.Now()
	// notify update
	a.peerUpdateChan <- true
}
