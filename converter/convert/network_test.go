package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph/simple"
	tu "utwente.nl/topology-to-dynetkat-coverter/test_util"
	"utwente.nl/topology-to-dynetkat-coverter/util"
	ug "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

func getMockEmptySwitch(nodeId int) *Switch {
	return &Switch{
		topoNode:   simple.Node(nodeId),
		controller: nil,
		flowTable:  getMockEmptyFT(),
	}
}

// returns n nodes
func getMockNodes(n int) []simple.Node {
	nodes := []simple.Node{}
	for i := range n {
		nodes = append(nodes, simple.Node(i))
	}
	return nodes
}

// returns the edges of a complete undirected graph with n nodes
func getMockEdges(nodes []simple.Node) []simple.WeightedEdge {
	edges := []simple.WeightedEdge{}
	for i, n1 := range nodes {
		for _, n2 := range nodes[i+1:] {
			edges = append(edges, simple.WeightedEdge{F: n1, T: n2, W: ug.DEFAULT_EDGE_WEIGHT})
		}
	}
	return edges
}

// makes a new undirected topology
func getMockTopology(n int) util.Graph {
	g := *ug.NewWeightedUndirectedGraph()
	nodes := getMockNodes(n)
	edges := getMockEdges(nodes)
	for _, node := range nodes {
		g.AddNode(node)
	}

	for _, edge := range edges {
		g.SetWeightedEdge(edge)
	}

	return g
}

func Test_mapNodeToSwitch(t *testing.T) {
	type args struct {
		switches []*Switch
	}

	pair := tu.GetOrderedMapPairFunc[int64, *Switch]()

	tests := []struct {
		name string
		args args
		want *om.OrderedMap[int64, *Switch]
	}{
		{
			name: "No Switches [Success]",
			args: args{
				switches: []*Switch{},
			},
			want: om.New[int64, *Switch](),
		},
		{
			name: "Non-Empty Switch Array [Success]",
			args: args{
				switches: []*Switch{
					getMockEmptySwitch(1),
					getMockEmptySwitch(-100),
					getMockEmptySwitch(348),
					getMockEmptySwitch(0),
				},
			},
			want: om.New[int64, *Switch](om.WithInitialData(
				pair(1, getMockEmptySwitch(1)),
				pair(-100, getMockEmptySwitch(-100)),
				pair(348, getMockEmptySwitch(348)),
				pair(0, getMockEmptySwitch(0)),
			)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mapNodeToSwitch(tt.args.switches)
			tu.AssertEqualMaps(t, tt.want, &got)

			// check if the map points to the same switch address as the array
			for i := 0; i < len(tt.args.switches); i++ {
				nodeId := tt.args.switches[i].topoNode.ID()
				gotSw, _ := got.Get(nodeId)
				assert.Same(t, tt.args.switches[i], gotSw)
			}
		})
	}
}

func Test_makeSwitchesFromTopology(t *testing.T) {
	type args struct {
		topo       util.Graph
		edgeToLink om.OrderedMap[util.I64Tup, *Link]
	}
	tests := []struct {
		name    string
		args    args
		want    []*Switch
		wantErr string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeSwitchesFromTopology(tt.args.topo, tt.args.edgeToLink)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func Test_makeLinks(t *testing.T) {
	type K = util.I64Tup
	type V = *Link

	pair := tu.GetOrderedMapPairFunc[K, V]()
	newMap := tu.GetOrderedMapFunc[K, V]()

	type args struct {
		topo   util.Graph
		portNr *int64
	}
	tests := []struct {
		name        string
		args        args
		wantErr     string
		assertSetup func(*testing.T, *om.OrderedMap[K, V], *int64)
	}{
		{
			name: "Nil Port Number [Validation Error]",
			args: args{
				topo: *ug.NewWeightedUndirectedGraph(), portNr: nil,
			},
			assertSetup: func(t *testing.T, linkMap *om.OrderedMap[K, V], portNr *int64) {
				tu.AssertEqualMaps(t, newMap(), linkMap)
			},
			wantErr: fmt.Sprintf(util.ErrNilArgument, "portNr"),
		},
		{
			name: "No edges [Success]",
			args: args{
				topo:   *ug.NewWeightedUndirectedGraph(),
				portNr: new(int64),
			},
			assertSetup: func(t *testing.T, linkMap *om.OrderedMap[K, V], portNr *int64) {
				assert.Equal(t, int64(0), *portNr)
				tu.AssertEqualMaps(t, om.New[K, V](), linkMap)
			},
		},
		{
			name: "No edges [Success]",
			args: args{
				topo:   *ug.NewWeightedUndirectedGraph(),
				portNr: new(int64),
			},
			assertSetup: func(t *testing.T, linkMap *om.OrderedMap[K, V], portNr *int64) {
				assert.Equal(t, int64(0), *portNr)
				tu.AssertEqualMaps(t, om.New[K, V](), linkMap)
			},
		},
		{
			name: "Non-Empty Graph [Success]",
			args: args{
				topo:   getMockTopology(3),
				portNr: new(int64),
			},
			assertSetup: func(t *testing.T, linkMap *om.OrderedMap[K, V], portNr *int64) {
				expected := newMap(
					pair(
						K{Fst: 0, Snd: 1},
						&Link{
							simple.WeightedEdge{F: simple.Node(0), T: simple.Node(1), W: 1},
							0,
							1,
						},
					),
					pair(
						K{Fst: 0, Snd: 2},
						&Link{
							simple.WeightedEdge{F: simple.Node(0), T: simple.Node(2), W: 1},
							2,
							3,
						},
					),
					pair(
						K{Fst: 1, Snd: 2},
						&Link{
							simple.WeightedEdge{F: simple.Node(1), T: simple.Node(2), W: 1},
							4,
							5,
						},
					),
				)

				assert.Equal(t, int64(6), *portNr)
				tu.AssertEqualMaps(t, expected, linkMap)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := makeLinks(tt.args.topo, tt.args.portNr)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
			if tt.assertSetup != nil {
				tt.assertSetup(t, &got, tt.args.portNr)
			}
		})
	}
}
