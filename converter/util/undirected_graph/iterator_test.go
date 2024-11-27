package undirectedgraph

import (
	"testing"

	"github.com/stretchr/testify/assert"
	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	tu "utwente.nl/topology-to-dynetkat-coverter/test_util"
)

func TestNewNodes(t *testing.T) {
	type args struct {
		nodes om.OrderedMap[int64, graph.Node]
	}
	type want struct {
		curr      *om.Pair[int64, graph.Node]
		pos       int
		length    int
		nextValue bool
		node      graph.Node
	}

	newMap := tu.GetOrderedMapFunc[int64, graph.Node]()
	pair := tu.GetOrderedMapPairFunc[int64, graph.Node]()
	pairP := tu.GetOrderedMapPairPointerFunc[int64, graph.Node]()

	tests := []struct {
		name  string
		steps uint
		args  args
		want  want
	}{
		{
			name: "Empty Map, 1 step [Success]",
			args: args{
				nodes: *newMap(),
			},
			steps: 1,
			want: want{
				curr:      nil,
				pos:       0,
				length:    0,
				nextValue: false,
				node:      nil,
			},
		},
		{
			name: "Non-Empty Map, 0 Steps [Success]",
			args: args{
				nodes: *newMap(
					pair(-5, simple.Node(-5)),
					pair(0, simple.Node(0)),
					pair(1, simple.Node(1)),
					pair(6, simple.Node(6)),
				),
			},
			steps: 0,
			want: want{
				curr:      nil,
				pos:       0,
				length:    4,
				nextValue: true,
				node:      nil,
			},
		},
		{
			name: "Non-Empty Map, 1 Step [Success]",
			args: args{
				nodes: *newMap(
					pair(-5, simple.Node(-5)),
					pair(0, simple.Node(0)),
					pair(1, simple.Node(1)),
					pair(6, simple.Node(6)),
				),
			},
			steps: 1,
			want: want{
				curr:      pairP(-5, simple.Node(-5)),
				pos:       1,
				length:    3,
				nextValue: true,
				node:      simple.Node(-5),
			},
		},
		{
			name: "Non-Empty Map, n-1 Steps [Success]",
			args: args{
				nodes: *newMap(
					pair(-5, simple.Node(-5)),
					pair(0, simple.Node(0)),
					pair(1, simple.Node(1)),
					pair(6, simple.Node(6)),
				),
			},
			steps: 3,
			want: want{
				curr:      pairP(1, simple.Node(1)),
				pos:       3,
				length:    1,
				nextValue: true,
				node:      simple.Node(1),
			},
		},
		{
			name: "Non-Empty Map, n Steps [Success]",
			args: args{
				nodes: *newMap(
					pair(-5, simple.Node(-5)),
					pair(0, simple.Node(0)),
					pair(1, simple.Node(1)),
					pair(6, simple.Node(6)),
				),
			},
			steps: 4,
			want: want{
				curr:      pairP(6, simple.Node(6)),
				pos:       4,
				length:    0,
				nextValue: true,
				node:      simple.Node(6),
			},
		},
		{
			name: "Non-Empty Map, More Steps Than Elements [Success]",
			args: args{
				nodes: *newMap(
					pair(-5, simple.Node(-5)),
					pair(0, simple.Node(0)),
					pair(1, simple.Node(1)),
					pair(6, simple.Node(6)),
				),
			},
			steps: 5,
			want: want{
				curr:      pairP(6, simple.Node(6)),
				pos:       4,
				length:    0,
				nextValue: false,
				node:      simple.Node(6),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aNext := false
			iter := NewNodes(tt.args.nodes)
			for range tt.steps {
				aNext = iter.Next()
			}

			assert.Equal(t, tt.want.pos, iter.pos)
			assert.Equal(t, tt.want.length, iter.Len())

			if tt.want.curr != nil {
				assert.Equal(t, tt.want.curr.Key, iter.curr.Key)
				assert.Equal(t, tt.want.curr.Value, iter.curr.Value)
				assert.Equal(t, tt.want.node, iter.Node())
			} else {
				assert.Nil(t, iter.curr)
			}

			// must be at the end because of Next which modifies iter.curr
			if tt.steps == 0 {
				assert.Equal(t, tt.want.nextValue, iter.Next())
			} else {
				assert.Equal(t, tt.want.nextValue, aNext)
			}
		})
	}
}

func TestNodesByWeightedEdge(t *testing.T) {
	type args struct {
		nodes om.OrderedMap[int64, graph.Node]
		edges om.OrderedMap[int64, graph.WeightedEdge]
	}
	type want struct {
		currEdge  *om.Pair[int64, graph.WeightedEdge]
		pos       int
		length    int
		nextValue bool
		node      graph.Node
	}

	nNewMap := tu.GetOrderedMapFunc[int64, graph.Node]()
	nPair := tu.GetOrderedMapPairFunc[int64, graph.Node]()
	eNewMap := tu.GetOrderedMapFunc[int64, graph.WeightedEdge]()
	ePair := tu.GetOrderedMapPairFunc[int64, graph.WeightedEdge]()
	ePairP := tu.GetOrderedMapPairPointerFunc[int64, graph.WeightedEdge]()

	cGraphNodes := func() om.OrderedMap[int64, graph.Node] {
		return *nNewMap(
			nPair(0, simple.Node(0)),
			nPair(1, simple.Node(1)),
			nPair(2, simple.Node(2)),
			nPair(3, simple.Node(3)),
			nPair(4, simple.Node(4)),
			nPair(5, simple.Node(5)),
		)
	}
	cGraphEdges := func() om.OrderedMap[int64, graph.WeightedEdge] {
		return *eNewMap(
			ePair(0, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(0)}),
			ePair(1, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(1)}),
			ePair(3, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(3)}),
			ePair(4, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(4)}),
			ePair(5, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(5)}),
		)
	}

	tests := []struct {
		name  string
		steps uint
		args  args
		want  want
	}{
		{
			name: "Empty Map, 1 Step [Success]",
			args: args{
				nodes: *nNewMap(),
				edges: *eNewMap(),
			},
			steps: 1,
			want: want{
				currEdge:  nil,
				pos:       0,
				length:    0,
				nextValue: false,
				node:      nil,
			},
		},
		{
			name: "Non-Empty Map, 0 Steps [Success]",
			args: args{ // assume a complete graph
				nodes: cGraphNodes(),
				edges: cGraphEdges(),
			},
			steps: 0,
			want: want{
				currEdge:  nil,
				pos:       0,
				length:    5,
				nextValue: true,
				node:      nil,
			},
		},
		{
			name: "Non-Empty Map, 1 Step [Success]",
			args: args{ // assume a complete graph
				nodes: cGraphNodes(),
				edges: cGraphEdges(),
			},
			steps: 1,
			want: want{
				currEdge:  ePairP(0, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(0)}),
				pos:       1,
				length:    4,
				nextValue: true,
				node:      simple.Node(0),
			},
		},
		{
			name: "Non-Empty Map, n-1 Steps [Success]",
			args: args{
				nodes: cGraphNodes(),
				edges: cGraphEdges(),
			},
			steps: 4,
			want: want{
				currEdge:  ePairP(4, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(4)}),
				pos:       4,
				length:    1,
				nextValue: true,
				node:      simple.Node(4),
			},
		},
		{
			name: "Non-Empty Map, n Steps [Success]",
			args: args{
				nodes: cGraphNodes(),
				edges: cGraphEdges(),
			},
			steps: 5,
			want: want{
				currEdge:  ePairP(5, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(5)}),
				pos:       5,
				length:    0,
				nextValue: true,
				node:      simple.Node(5),
			},
		},
		{
			name: "Non-Empty Map, More Steps Than Elements [Success]",
			args: args{
				nodes: cGraphNodes(),
				edges: cGraphEdges(),
			},
			steps: 6,
			want: want{
				currEdge:  ePairP(5, simple.WeightedEdge{F: simple.Node(2), T: simple.Node(5)}),
				pos:       5,
				length:    0,
				nextValue: false,
				node:      simple.Node(5),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aNext := false
			iter := NewNodesByWeightedEdge(tt.args.nodes, tt.args.edges)
			for range tt.steps {
				aNext = iter.Next()
			}

			assert.Equal(t, tt.want.pos, iter.pos)
			assert.Equal(t, tt.want.length, iter.Len())

			if tt.want.currEdge != nil {
				assert.Equal(t, tt.want.currEdge.Key, iter.currEdge.Key)
				assert.Equal(t, tt.want.currEdge.Value, iter.currEdge.Value)
				assert.Equal(t, tt.want.node, iter.Node())
			} else {
				assert.Nil(t, iter.currEdge)
			}

			// must be at the end because of Next which modifies iter.curr
			if tt.steps == 0 {
				assert.Equal(t, tt.want.nextValue, iter.Next())
			} else {
				assert.Equal(t, tt.want.nextValue, aNext)
			}
		})
	}
}
