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
)

var NETWORK_IDS []string = []string{
	"Atmnet.graphml",        // 21 nodes
	"Arpanet196912.graphml", // 4 nodes
	"Dataxchange.graphml",   // 6 nodes
	"Renam.graphml",         // 5 nodes
	"Netrail.graphml",       // 7 nodes
	"Getnet.graphml",        // 7 nodes
}
var NETWORK_ID string = NETWORK_IDS[1]

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

	ei, err := encode.NewEncodingInfo(network)
	if err != nil {
		log.Fatalln(err)
	}

	encoder := getEncoder("big-switch")
	fmtNet := encoder.Encode(ei)

	err = util.WriteToNewFile(OUTPUT_DIR, "output.txt", fmtNet)
	if err != nil {
		log.Println("Failed to write output file!")
		log.Printf("Error: %s\n", err.Error())
		return
	} else {
		log.Println("Done generating text file!")
	}

	err = util.WriteToNewPdf(OUTPUT_DIR, "output", fmtNet)
	if err != nil {
		log.Println("Failed to write output file!")
		log.Printf("Error: %s\n", err.Error())
		return
	} else {
		log.Println("Done generating PDF!")
	}
}

func getEncoder(encoderOption string) encode.NetworkEncoder {
	switch encoderOption {
	case "big-switch":
		return encode.NewLatexBigSwitchEncoder(false)
	case "simple":
		return encode.NewLatexSimpleEncoder(false)
	default:
		return encode.NewLatexBigSwitchEncoder(false)
	}
}
