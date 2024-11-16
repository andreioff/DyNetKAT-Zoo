package behavior

import (
	"errors"

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
		return errors.New("Nil network!")
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
		return errors.New("Nil newHost!")
	case n == nil:
		return errors.New("Nil network!")
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
	newEntries map[int64][]util.I64Tup,
) error {
	if n == nil {
		return errors.New("Nil network!")
	}

	for nodeId, portTups := range newEntries {
		sw := n.NodeIdToSw()[nodeId]
		c := sw.Controller()
		if c == nil {
			return errors.New("Switch has nil controller!")
		}

		c.AddNewFlowRules(nodeId, destHostId, portTups)
	}

	return nil
}
