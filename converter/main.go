package main

import (
	"log"

	"utwente.nl/topology-to-dynetkat-coverter/convert/encode"
	behavior "utwente.nl/topology-to-dynetkat-coverter/convert/network_behavior"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	DIR        = "../topologyzoo/sources/graphml/"
	OUTPUT_DIR = "./output/"
	HOSTS_NR   = 5
	// NETWORK_ID = "Atmnet.graphml" // 21 nodes
	NETWORK_ID = "Arpanet196912.graphml" // 4 nodes
)

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopos := util.ValidateTopologies(gs)

	log.Printf("Generating DyNetKAT encoding for topology with id: %s...\n", NETWORK_ID)
	topo, exists := validTopos[NETWORK_ID]

	if !exists {
		log.Fatalf("Topology with name '%s' is either invalid or does not exist\n", NETWORK_ID)
	}

	network, err := behavior.NewNetworkWithBehavior(
		topo,
		&behavior.OutsideHostConn{},
	)
	if err != nil {
		log.Fatalln(err)
	}

	encoder := encode.NewLatexSimpleEncoder(false)
	fmtNet, err := encoder.Encode(network)
	if err != nil {
		log.Fatalln(err)
	}

	util.WriteToNewFile(OUTPUT_DIR, "output.txt", fmtNet)
	log.Println("Done!")
}
