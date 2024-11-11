package main

import (
	"log"
	"maps"
	"os"
	"slices"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/convert/encode"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	DIR        = "../topologyzoo/sources/graphml/"
	OUTPUT_DIR = "./output/"
	FILE_PERM  = 0755
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

	n, err := convert.NewNetwork(topo)
	if err != nil {
		log.Fatalln(err)
	}

	err = n.AddAndConnectHosts(HOST_NR)
	if err != nil {
		log.Fatalln(err)
	}

	encoder := encode.NewLatexEncoder()
	fmtNet, err := encoder.Encode(n)
	if err != nil {
		log.Fatalln(err)
	}

	writeToFile("output.txt", fmtNet)
	log.Println("Done!")
}

func writeToFile(fileName, data string) {
	_, err := os.Stat(OUTPUT_DIR)
	if err != nil && !os.IsNotExist(err) {
		log.Fatalln(err)
	}

	if os.IsNotExist(err) {
		err = os.Mkdir(OUTPUT_DIR, FILE_PERM)
	}
	os.WriteFile(OUTPUT_DIR+fileName, []byte(data), FILE_PERM)
}
