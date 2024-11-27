package util

import (
	"fmt"
	"log"
	"os"
	"strings"

	om "github.com/wk8/go-ordered-map/v2"
	"github.com/yaricom/goGraphML/graphml"
	"gonum.org/v1/gonum/graph"
	ug "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

const (
	GRAPHML_EXT = ".graphml"
)

func getPathsFromDir(dirPath string) ([]string, error) {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		return []string{}, err
	}

	paths := []string{}
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), GRAPHML_EXT) {
			paths = append(paths, dirPath+file.Name())
		}
	}

	return paths, nil
}

func GetGraphMLs(dirPath string) ([]graphml.GraphML, error) {
	paths, err := getPathsFromDir(dirPath)
	if err != nil {
		log.Printf("Failed to read file paths!\n%s", err.Error())
		return []graphml.GraphML{}, err
	}

	graphs := []graphml.GraphML{}
	for _, path := range paths {
		r, err := os.Open(path)
		if err != nil {
			log.Printf("Failed to open %s! Skipping...", path)
			continue
		}

		fName := r.Name()[strings.LastIndex(r.Name(), "/")+1:]
		g := *graphml.NewGraphML(fName)
		err = g.Decode(r)
		if err != nil {
			log.Printf("Something went wrong while decoding %s.\n%s", fName, err.Error())
			continue
		}

		graphs = append(graphs, g)
	}

	return graphs, nil
}

func GraphMLToGraph(gml graphml.GraphML) (Graph, error) {
	if len(gml.Graphs) != 1 {
		return *ug.NewWeightedUndirectedGraph(), NewError(ErrGraphMLExactly1Graph)
	}

	gmlGraph := gml.Graphs[0]
	g := *ug.NewWeightedUndirectedGraph()

	gmlNodeToGNode := om.New[string, graph.Node]()
	for _, gmlNode := range gmlGraph.Nodes {
		newNode := g.NewNode()
		gmlNodeToGNode.Set(gmlNode.ID, newNode)
		g.AddNode(newNode)
	}

	for _, edge := range gmlGraph.Edges {
		from, to := edge.SourceNode().ID, edge.TargetNode().ID

		if from == to {
			log.Printf("%s: Skipping self-loop\n", gml.Description)
			continue
		}

		fromNode, _ := gmlNodeToGNode.Get(from)
		toNode, _ := gmlNodeToGNode.Get(to)
		g.SetWeightedEdge(g.NewWeightedEdge(fromNode, toNode, ug.DEFAULT_EDGE_WEIGHT))
	}

	return g, nil
}

func GraphMLsToGraphs(gmls []graphml.GraphML) om.OrderedMap[string, Graph] {
	gs := *om.New[string, Graph]()
	id := 0

	for _, gml := range gmls {
		g, err := GraphMLToGraph(gml)
		if err != nil {
			log.Printf(
				"Could not convert GraphML instace: %s! Skipping...\n%s.",
				gml.Description,
				err.Error(),
			)
			continue
		}

		if _, exists := gs.Get(gml.Description); exists {
			name := fmt.Sprintf("%s#%d", gml.Description, id)
			gs.Set(name, g)
			id++
			continue
		}

		gs.Set(gml.Description, g)
	}
	return gs
}
