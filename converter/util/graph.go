package util

import (
	"errors"
	"fmt"
	"log"

	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/topo"
)

type (
	Graph       = simple.UndirectedGraph
	NodeIdTuple = Tuple[int64, int64]
)

// return the valid topologies in the given array
func ValidateTopologies(tops []Graph) []Graph {
	validTops := []Graph{}

	for _, top := range tops {
		err := ValidateTopology(top)
		if err != nil {
			log.Printf("%s Skipping...", err)
			continue
		}

		validTops = append(validTops, top)
	}

	log.Printf("Processed %d topologies. %d were invalid.", len(tops), len(tops)-len(validTops))
	return validTops
}

// topologies must be connected
func ValidateTopology(top Graph) error {
	err := isConnected(top)
	if err != nil {
		return err
	}

	return nil
}

// verifies if the given graph is connected
func isConnected(top Graph) error {
	if top.Nodes().Len() == 0 {
		return errors.New("Empty graph!")
	}

	componentNr := len(topo.ConnectedComponents(&top))
	if componentNr > 1 {
		return errors.New(fmt.Sprintf("Disconnected graph with %d components!", componentNr))
	}

	return nil
}
