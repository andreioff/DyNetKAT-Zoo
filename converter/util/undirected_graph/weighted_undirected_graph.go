/*
Copyright ©2014 The Gonum Authors. All rights reserved.

Redistribution and use in source and binary forms, with or without
modification, are permitted provided that the following conditions are met:
    * Redistributions of source code must retain the above copyright
      notice, this list of conditions and the following disclaimer.
    * Redistributions in binary form must reproduce the above copyright
      notice, this list of conditions and the following disclaimer in the
      documentation and/or other materials provided with the distribution.
    * Neither the name of the Gonum project nor the names of its authors and
      contributors may be used to endorse or promote products derived from this
      software without specific prior written permission.

THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS" AND
ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
*/

package undirectedgraph

import (
	"fmt"

	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/iterator"
	"gonum.org/v1/gonum/graph/set/uid"
	"gonum.org/v1/gonum/graph/simple"
)

const DEFAULT_EDGE_WEIGHT = float64(1)

// Same implementation as "gonum.org/v1/gonum/graph/simple/WeightedUndirectedGraph",
// but adapted to use an ordered map
type WeightedUndirectedGraph struct {
	nodes om.OrderedMap[int64, graph.Node]
	edges om.OrderedMap[int64, om.OrderedMap[int64, graph.WeightedEdge]]

	self, absent float64

	nodeIDs *uid.Set
}

// NewWeightedUndirectedGraph returns an WeightedUndirectedGraph
func NewWeightedUndirectedGraph() *WeightedUndirectedGraph {
	return &WeightedUndirectedGraph{
		nodes: *om.New[int64, graph.Node](),
		edges: *om.New[int64, om.OrderedMap[int64, graph.WeightedEdge]](),

		self:   0,
		absent: 0,

		nodeIDs: uid.NewSet(),
	}
}

// AddNode adds n to the graph. It panics if the added node ID matches an existing node ID.
func (g *WeightedUndirectedGraph) AddNode(n graph.Node) {
	if _, exists := g.nodes.Get(n.ID()); exists {
		panic(fmt.Sprintf("simple: node ID collision: %d", n.ID()))
	}
	g.nodes.Set(n.ID(), n)
	g.nodeIDs.Use(n.ID())
}

// Edge returns the edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *WeightedUndirectedGraph) Edge(uid, vid int64) graph.Edge {
	return g.WeightedEdgeBetween(uid, vid)
}

// EdgeBetween returns the edge between nodes x and y.
func (g *WeightedUndirectedGraph) EdgeBetween(xid, yid int64) graph.Edge {
	return g.WeightedEdgeBetween(xid, yid)
}

// Edges returns all the edges in the graph.
func (g *WeightedUndirectedGraph) Edges() graph.Edges {
	if g.edges.Len() == 0 {
		return graph.Empty
	}
	var edges []graph.Edge
	for pair1 := g.edges.Oldest(); pair1 != nil; pair1 = pair1.Next() {
		xid, u := pair1.Key, pair1.Value
		for pair2 := u.Oldest(); pair2 != nil; pair2 = pair2.Next() {
			yid, e := pair2.Key, pair2.Value
			if yid < xid {
				// Do not consider edges when the To node ID is
				// before the From node ID. Both orientations
				// are stored.
				continue
			}
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedEdges(edges)
}

// From returns all nodes in g that can be reached directly from n.
func (g *WeightedUndirectedGraph) From(id int64) graph.Nodes {
	toEdges, exists := g.edges.Get(id)
	if !exists || toEdges.Len() == 0 {
		return graph.Empty
	}
	return NewNodesByWeightedEdge(g.nodes, toEdges)
}

// HasEdgeBetween returns whether an edge exists between nodes x and y.
func (g *WeightedUndirectedGraph) HasEdgeBetween(xid, yid int64) bool {
	toEdges, ok1 := g.edges.Get(xid)
	if !ok1 {
		return false
	}
	_, ok2 := toEdges.Get(yid)
	return ok2
}

// NewNode returns a new unique Node to be added to g. The Node's ID does
// not become valid in g until the Node is added to g.
func (g *WeightedUndirectedGraph) NewNode() graph.Node {
	if g.nodes.Len() == 0 {
		return simple.Node(0)
	}
	if int64(g.nodes.Len()) == uid.Max {
		panic("simple: cannot allocate node: no slot")
	}
	return simple.Node(g.nodeIDs.NewID())
}

// NewWeightedEdge returns a new weighted edge from the source to the destination node.
func (g *WeightedUndirectedGraph) NewWeightedEdge(
	from, to graph.Node,
	weight float64,
) graph.WeightedEdge {
	return simple.WeightedEdge{F: from, T: to, W: weight}
}

// Node returns the node with the given ID if it exists in the graph,
// and nil otherwise.
func (g *WeightedUndirectedGraph) Node(id int64) graph.Node {
	node, _ := g.nodes.Get(id)
	return node
}

// Nodes returns all the nodes in the graph.
//
// The returned graph.Nodes is only valid until the next mutation of
// the receiver.
func (g *WeightedUndirectedGraph) Nodes() graph.Nodes {
	if g.nodes.Len() == 0 {
		return graph.Empty
	}
	return NewNodes(g.nodes)
}

// NodeWithID returns a Node with the given ID if possible. If a graph.Node
// is returned that is not already in the graph NodeWithID will return true
// for new and the graph.Node must be added to the graph before use.
func (g *WeightedUndirectedGraph) NodeWithID(id int64) (n graph.Node, new bool) {
	n, ok := g.nodes.Get(id)
	if ok {
		return n, false
	}
	return simple.Node(id), true
}

// RemoveEdge removes the edge with the given end point IDs from the graph, leaving the terminal
// nodes. If the edge does not exist  it is a no-op.
func (g *WeightedUndirectedGraph) RemoveEdge(fid, tid int64) {
	if _, ok := g.nodes.Get(fid); !ok {
		return
	}
	if _, ok := g.nodes.Get(tid); !ok {
		return
	}

	fidToEdges := g.edges.GetPair(fid)
	if fidToEdges == nil {
		return
	}
	fidToEdges.Value.Delete(tid)

	tidToEdges := g.edges.GetPair(tid)
	if tidToEdges == nil {
		return
	}
	tidToEdges.Value.Delete(fid)
}

// RemoveNode removes the node with the given ID from the graph, as well as any edges attached
// to it. If the node is not in the graph it is a no-op.
func (g *WeightedUndirectedGraph) RemoveNode(id int64) {
	if _, ok := g.nodes.Get(id); !ok {
		return
	}
	g.nodes.Delete(id)

	fromEdges, exists := g.edges.Get(id)
	if exists {
		for pair := fromEdges.Oldest(); pair != nil; pair = pair.Next() {
			from := pair.Key
			toEdges := g.edges.GetPair(from)
			if toEdges != nil {
				toEdges.Value.Delete(id)
			}
		}
		g.edges.Delete(id)
	}

	g.nodeIDs.Release(id)
}

// SetWeightedEdge adds a weighted edge from one node to another. If the nodes do not exist, they are added
// and are set to the nodes of the edge otherwise.
// It will panic if the IDs of the e.From and e.To are equal.
func (g *WeightedUndirectedGraph) SetWeightedEdge(e graph.WeightedEdge) {
	var (
		from = e.From()
		fid  = from.ID()
		to   = e.To()
		tid  = to.ID()
	)

	if fid == tid {
		panic("simple: adding self edge")
	}

	if _, ok := g.nodes.Get(fid); !ok {
		g.AddNode(from)
	} else {
		g.nodes.Set(fid, from)
	}

	if _, ok := g.nodes.Get(tid); !ok {
		g.AddNode(to)
	} else {
		g.nodes.Set(tid, to)
	}

	if fm := g.edges.GetPair(fid); fm != nil {
		fm.Value.Set(tid, e)
	} else {
		fm := *om.New[int64, graph.WeightedEdge](om.WithInitialData(
			om.Pair[int64, graph.WeightedEdge]{Key: tid, Value: e},
		))
		g.edges.Set(fid, fm)
	}

	if tm := g.edges.GetPair(tid); tm != nil {
		tm.Value.Set(fid, e)
	} else {
		tm := *om.New[int64, graph.WeightedEdge](om.WithInitialData(
			om.Pair[int64, graph.WeightedEdge]{Key: fid, Value: e},
		))
		g.edges.Set(tid, tm)
	}
}

// Weight returns the weight for the edge between x and y if Edge(x, y) returns a non-nil Edge.
// If x and y are the same node or there is no joining edge between the two nodes the weight
// value returned is either the graph's absent or self value. Weight returns true if an edge
// exists between x and y or if x and y have the same ID, false otherwise.
func (g *WeightedUndirectedGraph) Weight(xid, yid int64) (w float64, ok bool) {
	if xid == yid {
		return g.self, true
	}
	if n, ok := g.edges.Get(xid); ok {
		if e, ok := n.Get(yid); ok {
			return e.Weight(), true
		}
	}
	return g.absent, false
}

// WeightedEdge returns the weighted edge from u to v if such an edge exists and nil otherwise.
// The node v must be directly reachable from u as defined by the From method.
func (g *WeightedUndirectedGraph) WeightedEdge(uid, vid int64) graph.WeightedEdge {
	return g.WeightedEdgeBetween(uid, vid)
}

// WeightedEdgeBetween returns the weighted edge between nodes x and y.
func (g *WeightedUndirectedGraph) WeightedEdgeBetween(xid, yid int64) graph.WeightedEdge {
	toEdges, ok1 := g.edges.Get(xid)
	if !ok1 {
		return nil
	}

	edge, ok2 := toEdges.Get(yid)
	if !ok2 {
		return nil
	}
	if edge.From().ID() == xid {
		return edge
	}
	return edge.ReversedEdge().(graph.WeightedEdge)
}

// WeightedEdges returns all the weighted edges in the graph.
func (g *WeightedUndirectedGraph) WeightedEdges() graph.WeightedEdges {
	if g.edges.Len() == 0 {
		return graph.Empty
	}
	var edges []graph.WeightedEdge
	for pair1 := g.edges.Oldest(); pair1 != nil; pair1 = pair1.Next() {
		xid, u := pair1.Key, pair1.Value
		for pair2 := u.Oldest(); pair2 != nil; pair2 = pair2.Next() {
			yid, e := pair2.Key, pair2.Value
			if yid < xid {
				// Do not consider edges when the To node ID is
				// before the From node ID. Both orientations
				// are stored.
				continue
			}
			edges = append(edges, e)
		}
	}
	if len(edges) == 0 {
		return graph.Empty
	}
	return iterator.NewOrderedWeightedEdges(edges)
}
