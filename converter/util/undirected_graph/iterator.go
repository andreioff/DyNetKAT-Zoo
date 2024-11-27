package undirectedgraph

import (
	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph"
)

// Nodes implements the graph.Nodes interfaces.
// The iteration order of Nodes is randomized.
type Nodes struct {
	nodes om.OrderedMap[int64, graph.Node]
	curr  *om.Pair[int64, graph.Node]
	pos   int
}

// NewNodes returns a Nodes initialized with the provided nodes, a
// map of node IDs to graph.Nodes. No check is made that the keys
// match the graph.Node IDs, and the map keys are not used.
//
// Behavior of the Nodes is unspecified if nodes is mutated after
// the call to NewNodes.
func NewNodes(nodes om.OrderedMap[int64, graph.Node]) *Nodes {
	return &Nodes{
		nodes: nodes,
		curr:  nil,
		pos:   0,
	}
}

// Len returns the remaining number of nodes to be iterated over.
func (n *Nodes) Len() int {
	return n.nodes.Len() - n.pos
}

// Next returns whether the next call of Node will return a valid node.
func (n *Nodes) Next() bool {
	if n.pos >= n.nodes.Len() {
		return false
	}

	if n.pos == 0 {
		n.curr = n.nodes.Oldest()
	} else {
		n.curr = n.curr.Next()
	}
	n.pos += 1
	return n.curr != nil
}

// Node returns the current node of the iterator. Next must have been
// called prior to a call to Node.
func (n *Nodes) Node() graph.Node {
	return n.curr.Value
}

// Reset returns the iterator to its initial state.
func (n *Nodes) Reset() {
	n.curr = nil
	n.pos = 0
}

// NodesByEdge implements the graph.Nodes interfaces.
// The iteration order of Nodes is randomized.
type NodesByEdge struct {
	nodes    om.OrderedMap[int64, graph.Node]
	edges    om.OrderedMap[int64, graph.WeightedEdge]
	currEdge *om.Pair[int64, graph.WeightedEdge]
	pos      int
}

// NewNodesByWeightedEdge returns a NodesByEdge initialized with the
// provided nodes, a map of node IDs to graph.Nodes, and the set
// of edges, a map of to-node IDs to graph.WeightedEdge, that can be
// traversed to reach the nodes that the NodesByEdge will iterate
// over. No check is made that the keys match the graph.Node IDs,
// and the map keys are not used.
//
// Behavior of the NodesByEdge is unspecified if nodes or edges
// is mutated after the call to NewNodes.
func NewNodesByWeightedEdge(
	nodes om.OrderedMap[int64, graph.Node],
	edges om.OrderedMap[int64, graph.WeightedEdge],
) *NodesByEdge {
	return &NodesByEdge{
		nodes:    nodes,
		edges:    edges,
		currEdge: nil,
		pos:      0,
	}
}

// Len returns the remaining number of nodes to be iterated over.
func (n *NodesByEdge) Len() int {
	return n.edges.Len() - n.pos
}

// Next returns whether the next call of Node will return a valid node.
func (n *NodesByEdge) Next() bool {
	if n.pos >= n.edges.Len() {
		return false
	}

	if n.pos == 0 {
		n.currEdge = n.edges.Oldest()
	} else {
		n.currEdge = n.currEdge.Next()
	}
	n.pos += 1

	return n.currEdge != nil
}

// Node returns the current node of the iterator. Next must have been
// called prior to a call to Node.
func (n *NodesByEdge) Node() graph.Node {
	node, _ := n.nodes.Get(n.currEdge.Key)
	return node
}

// Reset returns the iterator to its initial state.
func (n *NodesByEdge) Reset() {
	n.currEdge = nil
	n.pos = 0
}
