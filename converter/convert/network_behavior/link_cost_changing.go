package behavior

import (
	"log"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
	undirectedgraph "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

// TODO Note that in this case, some switches must receive a "drop all" policy update,
// which is not the case at the moment

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

		success, err := lcc.populateControllerNewFlowTables(n)
		if err != nil {
			return err
		}

		if success {
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

func (_ *LinkCostChanging) populateControllerNewFlowTables(n *convert.Network) (bool, error) {
	success := false
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
				return success, err
			}

			status, err := addEntriesToControllerNewFlowTables(n, h2.ID(), entries)
			if err != nil {
				return success, err
			}
			success = success || status
		}
	}
	return success, nil
}
