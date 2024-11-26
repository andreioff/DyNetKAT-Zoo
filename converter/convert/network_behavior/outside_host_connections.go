package behavior

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	HOSTS_NR         = 2
	OUTSIDE_HOSTS_NR = 1
	CONTROLLERS_NR   = 1
)

type OutsideHostConn struct{}

func (_ *OutsideHostConn) ModifyNetwork(n *convert.Network) error {
	err := n.AddAndConnectHosts(HOSTS_NR)
	if err != nil {
		return err
	}

	err = n.AddControllers(CONTROLLERS_NR)
	if err != nil {
		return err
	}

	newHosts, err := n.CreateHosts(OUTSIDE_HOSTS_NR)
	if err != nil {
		return err
	}

	populateControllerNewFlowTables(newHosts, n)
	return nil
}

func populateControllerNewFlowTables(newHosts []*convert.Host, n *convert.Network) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}

	for _, newHost := range newHosts {
		err := addNewHostConnFlowRules(newHost, n)
		if err != nil {
			return err
		}

	}
	return nil
}

func addNewHostConnFlowRules(newHost *convert.Host, n *convert.Network) error {
	switch {
	case newHost == nil:
		return util.NewError(util.ErrNilArgument, "newHost")
	case n == nil:
		return util.NewError(util.ErrNilArgument, "n")
	}

	for _, host := range n.Hosts() {
		newEntries, err := n.GetFlowRulesForSwitchPath(
			newHost.Switch(),
			host.Switch(),
			newHost.SwitchPort(),
			host.SwitchPort(),
		)
		if err != nil {
			return err
		}

		err = addEntriesToControllerNewFlowTables(n, host.ID(), newEntries)
		if err != nil {
			return err
		}
	}

	return nil
}

func addEntriesToControllerNewFlowTables(
	n *convert.Network,
	destHostId int64,
	newEntries om.OrderedMap[int64, []convert.FlowRule],
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

		c.AddNewFlowRules(nodeId, destHostId, frs)
	}

	return nil
}
