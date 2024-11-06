package convert

import (
	"gonum.org/v1/gonum/graph/path"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

/*
  Generate shortest path between every node in the topology
  Assign ports to each edge
  Randomly distribute hosts in the network
    Each host should be linked to a different port on the switch
  At every node, map the ports to their corresponding destinations (to be able to decide where to send each packet, this also means that the NetKAT policies will operate on 2 packet fields (for now))
  For each pair of hosts, use the computed shortest paths to generate NetKAT policies that encode the flow between these hosts
*/

type Network struct {
	topology      util.Graph
	shortestPaths path.AllShortest

	switches []Switch
	links    []Link

	portNr int64
}

func NewNetwork(topo util.Graph) *Network {
	var portNr int64 = 0

	links, portNr := makeLinksFromTopology(topo, portNr)
	return &Network{
		topology:      topo,
		shortestPaths: path.DijkstraAllPaths(&topo),
		switches:      makeSwitchesFromTopology(topo),
		links:         links,
		portNr:        portNr,
	}
}

func (n *Network) PortNr() int64 {
	return n.portNr
}

func makeSwitchesFromTopology(topo util.Graph) []Switch {
	switches := []Switch{}
	iter := topo.Nodes()
	for iter.Next() {
		switches = append(switches, *NewSwitch(iter.Node()))
	}

	return switches
}

func makeLinksFromTopology(topo util.Graph, portNr int64) ([]Link, int64) {
	links := []Link{}
	iter := topo.Edges()
	for iter.Next() {
		links = append(links, *NewLink(iter.Edge(), portNr, portNr+1))
		portNr += 2
	}

	return links, portNr
}
