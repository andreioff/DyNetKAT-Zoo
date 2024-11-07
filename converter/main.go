package main

import (
	"log"
	"slices"

	//"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var DIR = "../topologyzoo/sources/graphml/"

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopologies := util.ValidateTopologies(gs)

	// sort topologies in ascending order based on their nr of nodes and edges
	slices.SortFunc(validTopologies, util.GraphCmp)

	n, err := convert.NewNetwork(validTopologies[0])
	if err != nil {
		log.Fatalln(err)
	}

	err = n.AddAndConnectHosts(2)
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("\n%v\n", n)
}
