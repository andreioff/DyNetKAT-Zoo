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
	duplicateSwFT bool,
) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}

	for pair := newEntries.Oldest(); pair != nil; pair = pair.Next() {
		nodeId, frs := pair.Key, pair.Value
		sw, err := n.GetSwitch(nodeId)
		if err != nil {
			return err
		}

		c := sw.Controller()
		if c == nil {
			return util.NewError(util.ErrSwitchHasNilController)
		}

		err = c.AddNewFlowRules(nodeId, destHostId, frs, duplicateSwFT)
		if err != nil {
			return err
		}
	}

	return nil
}
