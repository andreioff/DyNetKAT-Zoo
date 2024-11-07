package util

import (
	"errors"
	"fmt"
	"log"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type (
	Graph       = simple.UndirectedGraph
	NodeIdTuple = Tuple[int64, int64]
)

// return the valid topologies in the given array
func ValidateTopologies(tops []Graph) []Graph {
	validTops := []Graph{}

	for _, top := range tops {
		err := ValidateTopology(top)
		if err != nil {
			log.Printf("%s Skipping...", err)
			continue
		}

		validTops = append(validTops, top)
	}

	log.Printf("Processed %d topologies. %d were invalid.", len(tops), len(tops)-len(validTops))
	return validTops
}

// topologies must be connected
func ValidateTopology(top Graph) error {
	err := isConnected(top)
	if err != nil {
		return err
	}

	return nil
}

// verifies if the given graph is connected
func isConnected(top Graph) error {
	if top.Nodes().Len() == 0 {
		return errors.New("Empty graph!")
	}

	componentNr := len(topo.ConnectedComponents(&top))
	if componentNr > 1 {
		return errors.New(fmt.Sprintf("Disconnected graph with %d components!", componentNr))
	}

	return nil
}

func GetNodesArrayFromIter(g Graph) []graph.Node {
	iter := g.Nodes()
	switches := []graph.Node{}

	for iter.Next() {
		switches = append(switches, iter.Node())
	}

	return switches
}

/*
returns -1 if  a < b
returns 0 if a = b
returns 1 if a > b
*/
func GraphCmp(a Graph, b Graph) int {
	aNodes := a.Nodes().Len()
	bNodes := b.Nodes().Len()
	aEdges := a.Edges().Len()
	bEdges := b.Edges().Len()

	if aNodes == bNodes && aEdges == bEdges {
		return 0
	}

	if aNodes < bNodes ||
		(aNodes == bNodes && aEdges < bEdges) {
		return -1
	}

	return 1
}

func GetIncidentEdges(g Graph, n graph.Node) ([]graph.Edge, error) {
	if n == nil {
		return []graph.Edge{}, errors.New("Node is nil!")
	}

	if g.Node(n.ID()) == nil {
		return []graph.Edge{}, errors.New("Node is not part of the graph!")
	}

	incidentEdges := []graph.Edge{}
	iter := g.Edges()

	for iter.Next() {
		if iter.Edge().To().ID() == n.ID() || iter.Edge().From().ID() == n.ID() {
			incidentEdges = append(incidentEdges, iter.Edge())
		}
	}
	return incidentEdges, nil
}
