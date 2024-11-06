package main

import (
	"log"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var DIR = "../topologyzoo/sources/graphml/"

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Panicf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopologies := util.ValidateTopologies(gs)

	net := convert.NewNetwork(validTopologies[0])
	log.Printf("%d", net.PortNr())
}
