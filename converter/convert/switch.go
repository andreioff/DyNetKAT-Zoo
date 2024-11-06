package convert

import "gonum.org/v1/gonum/graph"

type Switch struct {
	topoNode graph.Node
	hosts    []Host
}

func NewSwitch(node graph.Node) *Switch {
	return &Switch{
		topoNode: node,
		hosts:    []Host{},
	}
}

func (s *Switch) TopoNode() graph.Node {
	return s.topoNode
}

func (s *Switch) Hosts() []Host {
	return s.hosts
}
