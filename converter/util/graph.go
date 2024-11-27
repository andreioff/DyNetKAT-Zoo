package util

import (
	"log"

	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/topo"
	ug "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

type (
	Graph = ug.WeightedUndirectedGraph
)

// return the valid topologies in the given array
func ValidateTopologies(
	tops om.OrderedMap[string, Graph],
) om.OrderedMap[string, Graph] {
	validTops := *om.New[string, Graph]()

	for validTop := tops.Oldest(); validTop != nil; validTop = validTop.Next() {
		name, top := validTop.Key, validTop.Value
		err := ValidateTopology(top)
		if err != nil {
			log.Printf("%s: %s Skipping...", name, err)
			continue
		}

		validTops.Set(name, top)
	}

	log.Printf("Processed %d topologies. %d were invalid.", tops.Len(), tops.Len()-validTops.Len())
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
		return NewError(ErrEmptyGraph)
	}

	componentNr := len(topo.ConnectedComponents(&top))
	if componentNr > 1 {
		return NewError(ErrDisconnGraph, componentNr)
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
		return []graph.Edge{}, NewError(ErrNilArgument, "n")
	}

	if g.Node(n.ID()) == nil {
		return []graph.Edge{}, NewError(ErrNodeNotInGraph)
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
