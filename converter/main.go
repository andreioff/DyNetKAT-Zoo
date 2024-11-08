package main

import (
	"log"
	"slices"

	//"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var (
	DIR     = "../topologyzoo/sources/graphml/"
	HOST_NR = 5
)

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopologies := util.ValidateTopologies(gs)

	// sort topologies in ascending order based on their nr of nodes and edges
	slices.SortFunc(validTopologies, util.GraphCmp)

	networks := []*convert.Network{}
	for _, topo := range validTopologies {
		n, err := convert.NewNetwork(topo)
		if err != nil {
			log.Fatalln(err)
		}

		err = n.AddAndConnectHosts(5)
		if err != nil {
			log.Fatalln(err)
		}

		networks = append(networks, n)
	}

	log.Printf("\n%v\n", networks[0])
}
