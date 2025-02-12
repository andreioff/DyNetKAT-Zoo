package behavior

import (
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

/*
Given an empty network, it establishes the initial state of the network
by populating the flow tables of the corresponding switches to connect
a given number of hosts with each other. From this initial state,
the controllers receive new flow tables for the switches in the network
to establish connections between new, unconnected hosts and the existing
ones in the network.
*/
type OutsideHostConn struct{}

func (ohc *OutsideHostConn) ModifyNetwork(n *convert.Network, config BehaviorConfig) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}

	err := n.AddAndConnectHosts(config.Hosts_nr)
	if err != nil {
		return err
	}

	err = n.AddControllers(config.Controllers_nr)
	if err != nil {
		return err
	}

	newHosts, err := n.CreateHosts(config.Outside_hosts_nr)
	if err != nil {
		return err
	}

	return ohc.populateControllerNewFlowTables(newHosts, n)
}

func (ohc *OutsideHostConn) populateControllerNewFlowTables(
	newHosts []*convert.Host,
	n *convert.Network,
) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}

	for _, newHost := range newHosts {
		err := ohc.addNewHostConnFlowRules(newHost, n)
		if err != nil {
			return err
		}

	}
	return nil
}

func (ohc *OutsideHostConn) addNewHostConnFlowRules(
	newHost *convert.Host,
	n *convert.Network,
) error {
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

		err = addEntriesToControllerNewFlowTables(n, host.ID(), newEntries, true)
		if err != nil {
			return err
		}
	}

	return nil
}
