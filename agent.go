package element

import (
	fmt "fmt"
	"sync"
	"time"

	"github.com/containerd/typeurl"
	"github.com/gogo/protobuf/types"
	"github.com/hashicorp/memberlist"
	"github.com/pkg/errors"
)

const (
	defaultInterval      = time.Second * 10
	nodeReconcileTimeout = defaultInterval * 3
	nodeUpdateTimeout    = defaultInterval / 2
)

var (
	ErrUnknownConnectionType = errors.New("unknown connection type")
)

// Agent represents the node agent
type Agent struct {
	*subscribers

	config             *Config
	members            *memberlist.Memberlist
	peerUpdateChan     chan bool
	nodeEventChan      chan *NodeEvent
	registeredServices map[string]struct{}
	memberConfig       *memberlist.Config
	metadata           *Metadata
	messageHandlers    map[string]func(interface{}) error
}

// NewAgent returns a new node agent
func NewAgent(cfg *Config) (*Agent, error) {
	var (
		updateCh    = make(chan bool, 64)
		nodeEventCh = make(chan *NodeEvent, 64)
	)
	a := &Agent{
		subscribers:     newSubscribers(),
		config:          cfg,
		peerUpdateChan:  updateCh,
		nodeEventChan:   nodeEventCh,
		metadata:        &Metadata{},
		messageHandlers: make(map[string]func(interface{}) error),
	}
	mc, err := cfg.memberListConfig(a)
	if err != nil {
		return nil, err
	}
	ml, err := memberlist.Create(mc)
	if err != nil {
		return nil, err
	}
	a.members = ml
	a.memberConfig = mc

	return a, nil
}

// SyncInterval returns the cluster sync interval
func (a *Agent) SyncInterval() time.Duration {
	return a.memberConfig.PushPullInterval
}

// Update updates the agent payload
func (a *Agent) Update(payload *types.Any) {
	a.metadata.Payload = payload
}

// Send sends a message to the specified node
// **Note: `v` must be registered (typeurl.Register)
func (a *Agent) Send(n *Node, v interface{}) error {
	any, err := typeurl.MarshalAny(v)
	if err != nil {
		return err
	}
	data, err := any.Marshal()
	if err != nil {
		return err
	}
	return a.members.SendReliable(n.Node, data)
}

// RegisterMessageHandler registers a handler func for the specified type url for cluster messages
func (a *Agent) RegisterMessageHandler(url string, f func(v interface{}) error) error {
	if _, exists := a.messageHandlers[url]; exists {
		return fmt.Errorf("message handler exists for url %s", url)
	}
	a.messageHandlers[url] = f
	return nil
}

func newSubscribers() *subscribers {
	return &subscribers{
		subs: make(map[chan *NodeEvent]struct{}),
	}
}

type subscribers struct {
	mu   sync.Mutex
	subs map[chan *NodeEvent]struct{}
}

// Subscribe subscribes to the node event channel
func (s *subscribers) Subscribe() chan *NodeEvent {
	ch := make(chan *NodeEvent, 64)
	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()
	return ch
}

// Unsubscribe removes the channel from node events
func (s *subscribers) Unsubscribe(ch chan *NodeEvent) {
	s.mu.Lock()
	delete(s.subs, ch)
	s.mu.Unlock()
}

func (s *subscribers) send(e *NodeEvent) {
	s.mu.Lock()
	for ch := range s.subs {
		// non-blocking send
		select {
		case ch <- e:
		default:
		}
	}
	s.mu.Unlock()
}
