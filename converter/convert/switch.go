package convert

import (
	"gonum.org/v1/gonum/graph"
)

type Switch struct {
	topoNode  graph.Node
	hosts     []*Host
	destTable map[int64]map[int64]int64

	links []Link // outgoing links
}

func NewSwitch(node graph.Node, links []Link) *Switch {
	return &Switch{
		topoNode: node,
		hosts:    []*Host{},
		destTable: make(
			map[int64]map[int64]int64,
		), // maps host destination id and incoming port to outgoing port
		links: links,
	}
}

func (s *Switch) TopoNode() graph.Node {
	return s.topoNode
}

func (s *Switch) Hosts() []*Host {
	return s.hosts
}

func (s *Switch) DestTable() map[int64]map[int64]int64 {
	return s.destTable
}

func (s *Switch) AddHost(h *Host) {
	s.hosts = append(s.hosts, h)
}

func (s *Switch) FindLink(otherNodeId int64) *Link {
	for _, link := range s.links {
		if link.topoEdge.From().ID() == otherNodeId || link.topoEdge.To().ID() == otherNodeId {
			return &link
		}
	}
	return nil
}
