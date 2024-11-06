package convert

import "gonum.org/v1/gonum/graph"

type Link struct {
	topoEdge graph.Edge
	fromPort int64
	toPort   int64
}

func NewLink(edge graph.Edge, fromPort, toPort int64) *Link {
	return &Link{
		topoEdge: edge,
		fromPort: fromPort,
		toPort:   toPort,
	}
}

func (l *Link) TopoEdge() graph.Edge {
	return l.topoEdge
}

func (l *Link) FromPort() int64 {
	return l.fromPort
}

func (l *Link) ToPort() int64 {
	return l.toPort
}
