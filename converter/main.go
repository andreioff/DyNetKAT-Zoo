package main

import (
	"fmt"
	"log"

	"utwente.nl/topology-to-dynetkat-coverter/convert/encode"
	je "utwente.nl/topology-to-dynetkat-coverter/convert/encode/json_encoder_whole_network_updates"
	behavior "utwente.nl/topology-to-dynetkat-coverter/convert/network_behavior"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	DIR        = "../topologyzoo/sources/graphml"
	OUTPUT_DIR = "./output"
)

func main() {
	graphMLs, err := util.ReadGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to read graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopos := util.ValidateTopologies(gs)
	fmt.Println()
	seed := int64(15)

	for pair := validTopos.Oldest(); pair != nil; pair = pair.Next() {
		generateEncoding(pair.Key, pair.Value, seed)
	}
}

func generateEncoding(topoName string, topo util.Graph, seed int64) {
	log.Printf("Generating DyNetKAT encoding for topology with id: %s...\n", topoName)
	config := behavior.BehaviorConfig{
		Hosts_nr:         0,
		Outside_hosts_nr: uint(topo.Nodes().Len() / 2),
		Controllers_nr:   2,
	}
	if topo.Nodes().Len() < 10 {
		config.Outside_hosts_nr = uint(topo.Nodes().Len())
	}
	util.SetRandGenSeed(seed)

	network, err := behavior.NewNetworkWithBehavior(
		topo,
		behavior.NewPairwiseHostConn(true),
		config,
	)
	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}

	ei, err := encode.NewEncodingInfo(network)
	if err != nil {
		log.Fatalf("Error: %s\n", err.Error())
	}

	fmtNet, err := je.NewJsonEncoder().Encode(ei)
	if err != nil {
		log.Println("Failed to encode data!")
		log.Printf("Error: %s\n", err.Error())
	}

	err = util.WriteToNewFile(
		OUTPUT_DIR,
		fmt.Sprintf(
			"output_l%d_%s_s%d_h%d_oh%d_c%d.json",
			topo.Nodes().Len(),
			topoName,
			seed,
			config.Hosts_nr,
			config.Outside_hosts_nr,
			config.Controllers_nr,
		),
		fmtNet,
	)
	if err != nil {
		log.Println("Failed to write output file!")
		log.Printf("Error: %s\n", err.Error())
		return
	} else {
		log.Println("Done generating text file!")
		fmt.Println()
	}
}
