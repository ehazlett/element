package element

import "github.com/hashicorp/memberlist"

type Node struct {
	*memberlist.Node
}

// Nodes returns all known nodes in the cluster
func (a *Agent) Nodes() ([]*Node, error) {
	nodes := []*Node{}
	for _, node := range a.members.Members() {
		nodes = append(nodes, &Node{node})
	}
	return nodes, nil
}

// Self returns the local peer information
func (a *Agent) Self() *Node {
	self := a.members.LocalNode()
	return &Node{self}
}
