package convert

import (
	"gonum.org/v1/gonum/graph"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Switch struct {
	topoNode   graph.Node
	controller *Controller
	hosts      []*Host
	flowTable  *FlowTable

	links []*Link // outgoing links
}

func NewSwitch(node graph.Node, links []*Link) (*Switch, error) {
	if node == nil {
		return &Switch{}, util.NewError(util.ErrNilArgument, "node")
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

/*
Returns, in the correct order, the ports of the link between this switch and the given switch id.
Returns an error if the link could not be found.
*/
func (s *Switch) GetLinkPorts(otherNodeId int64) (int64, int64, error) {
	for _, link := range s.links {
		if link.topoEdge.From().ID() == otherNodeId {
			return link.ToPort(), link.FromPort(), nil
		}
		if link.topoEdge.To().ID() == otherNodeId {
			return link.FromPort(), link.ToPort(), nil
		}
	}
	return 0, 0, util.NewError(util.ErrNoLinkBetweenSwitches)
}

func (s *Switch) GetController() *Controller {
	return s.controller
}

func (s *Switch) SetController(c *Controller) {
	s.controller = c
}
