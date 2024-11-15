package behavior

import (
	"errors"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	HOSTS_NR         = 2
	OUTSIDE_HOSTS_NR = 2
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
		ft, err := findOrCreateNewFlowTable(n.NodeIdToSw()[nodeId])
		if err != nil {
			return err
		}

		for _, inPortOutPort := range portTups {
			ft.AddEntry(destHostId, inPortOutPort.Fst, inPortOutPort.Snd)
		}
	}

	return nil
}

// Returns the new flow table of the given switch that will be installed
// by the associated controller. If the controller does not have this new flow table,
// it creates one by coping over the table of the given switch.
func findOrCreateNewFlowTable(sw *convert.Switch) (*convert.FlowTable, error) {
	if sw == nil {
		return convert.NewFlowTable(), errors.New("Nil switch!")
	}

	c := sw.Controller()
	if c == nil {
		return convert.NewFlowTable(), errors.New("Switch has nil controller!")
	}

	nodeId := sw.TopoNode().ID()
	ft, exists := c.NewFlowTables()[nodeId]
	if exists {
		return ft, nil
	}

	err := c.CopyFlowTable(nodeId)
	if err != nil {
		return convert.NewFlowTable(), err
	}
	return c.NewFlowTables()[nodeId], nil
}
