package behavior

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	HOSTS_NR         = 3
	OUTSIDE_HOSTS_NR = 1
	CONTROLLERS_NR   = 1
)

type Behavior interface {
	ModifyNetwork(n *convert.Network) error
}

func NewNetworkWithBehavior(topo util.Graph, b Behavior) (*convert.Network, error) {
	newNet, err := convert.NewNetwork(topo)
	if err != nil {
		return newNet, err
	}

	net := *newNet              // copy the value at the pointer's memory location
	err = b.ModifyNetwork(&net) // apply the modification on the copy
	if err != nil {
		// if something goes bad, return the initial, empty network
		return newNet, err
	}
	return &net, nil
}

func addEntriesToControllerNewFlowTables(
	n *convert.Network,
	destHostId int64,
	newEntries om.OrderedMap[int64, []convert.FlowRule],
) (bool, error) {
	if n == nil {
		return false, util.NewError(util.ErrNilArgument, "n")
	}

	success := false
	for pair := newEntries.Oldest(); pair != nil; pair = pair.Next() {
		nodeId, frs := pair.Key, pair.Value
		sw, err := n.GetSwitch(nodeId)
		if err != nil {
			return success, err
		}

		c := sw.Controller()
		if c == nil {
			return success, util.NewError(util.ErrSwitchHasNilController)
		}

		status, err := c.AddNewFlowRules(nodeId, destHostId, frs)
		if err != nil {
			return success, err
		}
		success = success || status
	}

	return success, nil
}
