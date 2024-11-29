package behavior

import (
	"log"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
	undirectedgraph "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

const (
	CHANGING_COSTS_ATTEMPTS = 10
	MAX_LINK_COST           = 20
	MIN_LINK_COST           = int(undirectedgraph.DEFAULT_EDGE_WEIGHT)
)

type LinkCostChanging struct{}

func (lcc *LinkCostChanging) ModifyNetwork(n *convert.Network) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}

	err := n.AddAndConnectHosts(HOSTS_NR)
	if err != nil {
		return err
	}

	err = n.AddControllers(CONTROLLERS_NR)
	if err != nil {
		return err
	}

	return lcc.tryPopulateNewFlowTables(n)
}

// Returns an error if something goes wrong during the process, or if
// the maximum no. of attempts is reached.
func (lcc *LinkCostChanging) tryPopulateNewFlowTables(n *convert.Network) error {
	for i := range CHANGING_COSTS_ATTEMPTS {
		randomCosts := util.RandomInts(n.TopoEdgesLen(), MIN_LINK_COST, MAX_LINK_COST)
		err := n.ModifyLinkCosts(randomCosts)
		if err != nil {
			return err
		}

		err = lcc.populateControllerNewFlowTables(n)
		if err != nil {
			return err
		}

		allNewFTsRemoved, err := lcc.removeDuplicateFTs(n)
		if err != nil {
			return err
		}

		if !allNewFTsRemoved {
			return nil
		}

		log.Printf(
			"Failed to generate new host communication paths. Retrying... [%d/%d]",
			i+1,
			CHANGING_COSTS_ATTEMPTS,
		)
	}

	return util.NewError(util.ErrMaxCostChangingAttemptsReached, CHANGING_COSTS_ATTEMPTS)
}

func (_ *LinkCostChanging) populateControllerNewFlowTables(n *convert.Network) error {
	for i := range len(n.Hosts()) {
		for j := range len(n.Hosts()) {
			if i == j {
				continue
			}
			h1, h2 := n.Hosts()[i], n.Hosts()[j]
			entries, err := n.GetFlowRulesForSwitchPath(
				h1.Switch(),
				h2.Switch(),
				h1.SwitchPort(),
				h2.SwitchPort(),
			)
			if err != nil {
				return err
			}

			err = addEntriesToControllerNewFlowTables(n, h2.ID(), entries, false)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Removes any new flow tables that have the same entries as the corresponding switch flow table.
// Returns true if all new flow tables were removed, and false otherwise.
func (_ *LinkCostChanging) removeDuplicateFTs(n *convert.Network) (bool, error) {
	if n == nil {
		return false, util.NewError(util.ErrNilArgument, "n")
	}

	allRemoved := true
	for _, sw := range n.Switches() {
		c := sw.Controller()
		if c == nil {
			return false, util.NewError(util.ErrSwitchHasNilController)
		}

		newFt, newFtExists := c.NewFlowTables().Get(sw.TopoNode().ID())
		if newFtExists && newFt.IsEqual(sw.FlowTable()) {
			c.NewFlowTables().Delete(sw.TopoNode().ID())
		} else if newFtExists {
			allRemoved = false
		}
	}

	return allRemoved, nil
}
