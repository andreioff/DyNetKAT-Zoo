package main

import (
	"log"
	"os"
	//"slices"

	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/convert/encode"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const (
	DIR        = "../topologyzoo/sources/graphml/"
	OUTPUT_DIR = "./output/"
	FILE_PERM  = 0755
	HOST_NR    = 5
	INDEX      = 2 // 23 switches
)

func main() {
	graphMLs, err := util.GetGraphMLs(DIR)
	if err != nil {
		log.Fatalf("Failed to load graphs from directory: %s\n%s", DIR, err.Error())
	}

	gs := util.GraphMLsToGraphs(graphMLs)
	validTopologies := util.ValidateTopologies(gs)

	// sort topologies in ascending order based on their nr of nodes and edges
	// slices.SortFunc(validTopologies, util.GraphCmp)

	networks := []*convert.Network{}
	for _, topo := range validTopologies {
		n, err := convert.NewNetwork(topo)
		if err != nil {
			log.Fatalln(err)
		}

		err = n.AddAndConnectHosts(HOST_NR)
		if err != nil {
			log.Fatalln(err)
		}

		networks = append(networks, n)
	}

	encoder := encode.NewLatexEncoder()
	fmtNet, err := encoder.Encode(networks[INDEX])
	if err != nil {
		log.Fatalln(err)
	}

	writeToFile("output.txt", fmtNet)
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
