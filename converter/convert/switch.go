package convert

import (
	"gonum.org/v1/gonum/graph"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Switch struct {
	topoNode  graph.Node
	hosts     []*Host
	destTable map[util.I64Tup][]int64
	// maps host destination id and incoming port to outgoing port

	links []Link // outgoing links
}

func NewSwitch(node graph.Node, links []Link) *Switch {
	return &Switch{
		topoNode:  node,
		hosts:     []*Host{},
		destTable: make(map[util.I64Tup][]int64),
		links:     links,
	}
}

func (s *Switch) TopoNode() graph.Node {
	return s.topoNode
}

func (s *Switch) Hosts() []*Host {
	return s.hosts
}

func (s *Switch) DestTable() map[util.I64Tup][]int64 {
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

func (s *Switch) AddDestEntry(destHostId, inPort, outPort int64) {
	key := util.NewI64Tup(destHostId, inPort)

	// do not add duplicate entries
	if s.hasEntry(key, outPort) {
		return
	}

	s.destTable[key] = append(s.destTable[key], outPort)
}

func (s *Switch) hasEntry(key util.I64Tup, value int64) bool {
	if _, exists := s.destTable[key]; !exists {
		return false
	}

	for _, v := range s.destTable[key] {
		if v == value {
			return true
		}
	}

	return false
}
