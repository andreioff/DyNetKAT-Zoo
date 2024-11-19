package convert

import (
	"gonum.org/v1/gonum/graph"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Switch struct {
	topoNode   graph.Node
	controller *Controller
	flowTable  *FlowTable

	links []*Link // outgoing links
}

func NewSwitch(node graph.Node, links []*Link) (*Switch, error) {
	if node == nil {
		return &Switch{}, util.NewError(util.ErrNilArgument, "node")
	}

	if !onlyIncidentLinks(node, links) {
		return &Switch{}, util.NewError(util.ErrOnlyIncidentLinksForSwitch)
	}

	return &Switch{
		topoNode:   node,
		controller: nil,
		flowTable:  NewFlowTable(),
		links:      links,
	}, nil
}

func onlyIncidentLinks(node graph.Node, links []*Link) bool {
	if node == nil {
		return false
	}

	for _, l := range links {
		if !l.IsIncidentToNode(node.ID()) {
			return false
		}
	}
	return true
}

func (s *Switch) TopoNode() graph.Node {
	return s.topoNode
}

func (s *Switch) FlowTable() *FlowTable {
	return s.flowTable
}

func (s *Switch) Controller() *Controller {
	return s.controller
}

func (s *Switch) SetController(c *Controller) {
	s.controller = c
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
