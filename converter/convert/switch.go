package convert

import (
	"errors"

	"gonum.org/v1/gonum/graph"
)

type Switch struct {
	topoNode   graph.Node
	controller *Controller
	hosts      []*Host
	flowTable  *FlowTable

	links []Link // outgoing links
}

func NewSwitch(node graph.Node, links []Link) (*Switch, error) {
	if node == nil {
		return &Switch{}, errors.New("Nil topology node!")
	}

	return &Switch{
		topoNode:   node,
		hosts:      []*Host{},
		controller: nil,
		flowTable:  NewFlowTable(),
		links:      links,
	}, nil
}

func (s *Switch) TopoNode() graph.Node {
	return s.topoNode
}

func (s *Switch) Hosts() []*Host {
	return s.hosts
}

func (s *Switch) FlowTable() *FlowTable {
	return s.flowTable
}

func (s *Switch) Controller() *Controller {
	return s.controller
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

func (s *Switch) GetController() *Controller {
	return s.controller
}

func (s *Switch) SetController(c *Controller) {
	s.controller = c
}
