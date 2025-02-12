package behavior

import (
	"gonum.org/v1/gonum/graph"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	ug "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

/*
Topology with form:

		 1 -- 3
		/      \
	 0        5
		\      /
		 2 -- 4
*/
func getMockTopology1() ug.WeightedUndirectedGraph {
	g := *ug.NewWeightedUndirectedGraph()
	edges := []struct{ from, to int }{
		{0, 1},
		{0, 2},
		{1, 3},
		{2, 4},
		{3, 5},
		{4, 5},
	}

	idToNode := make(map[int]graph.Node)
	for i := range 6 {
		newNode := g.NewNode()
		idToNode[i] = newNode
		g.AddNode(newNode)
	}

	for _, edge := range edges {
		fromNode := idToNode[edge.from]
		toNode := idToNode[edge.to]
		g.SetWeightedEdge(g.NewWeightedEdge(fromNode, toNode, ug.DEFAULT_EDGE_WEIGHT))
	}
	return g
}

func getMockEmptyNetwork1() *convert.Network {
	n, err := convert.NewNetwork(getMockTopology1())
	if err != nil {
		panic("Failed to create mock:\n\t" + err.Error())
	}

	return n
}
