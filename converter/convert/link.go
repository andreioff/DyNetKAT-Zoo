package convert

import (
	"gonum.org/v1/gonum/graph"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Link struct {
	topoEdge graph.Edge
	fromPort int64
	toPort   int64
}

func NewLink(edge graph.Edge, fromPort, toPort int64) (*Link, error) {
	if edge == nil {
		return &Link{}, util.NewError(util.ErrNilArgument, "edge")
	}

	return &Link{
		topoEdge: edge,
		fromPort: fromPort,
		toPort:   toPort,
	}, nil
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

func (l *Link) IsIncidentToNode(nodeId int64) bool {
	return l.topoEdge.From().ID() == nodeId || l.topoEdge.To().ID() == nodeId
}
