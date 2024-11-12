package main

import (
	"log"
	"maps"
	"slices"

	"utwente.nl/topology-to-dynetkat-coverter/convert/encode"
	behavior "utwente.nl/topology-to-dynetkat-coverter/convert/network_behavior"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	DIR        = "../topologyzoo/sources/graphml/"
	OUTPUT_DIR = "./output/"
	HOST_NR    = 5
	NETWORK_ID = "Atmnet.graphml" // 21 switches
)

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	util.ValidateTopologies(slices.Collect(maps.Values(gs)))

	log.Printf("Processing network with topology id: %s", NETWORK_ID)
	topo := gs[NETWORK_ID]

	if util.ValidateTopology(topo) != nil {
		log.Fatalln("Invalid network!")
	}

	network, err := behavior.NewNetworkWithBehavior(topo, &behavior.OutsideHostConn{})
	if err != nil {
		log.Fatalln(err)
	}

	encoder := encode.NewLatexEncoder()
	fmtNet, err := encoder.Encode(network)
	if err != nil {
		log.Fatalln(err)
	}

	util.WriteToNewFile(OUTPUT_DIR, "output.txt", fmtNet)
	log.Println("Done!")
}
