package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

func TestConvert_NewLink(t *testing.T) {
	var simpleEdge graph.Edge = simple.Edge{F: simple.Node(0), T: simple.Node(0)}

	cases := map[string]struct {
		edge        graph.Edge
		fromPort    int64
		toPort      int64
		assertSetup func(*testing.T, *Link, error)
	}{
		"Nil edge [Validation error]": {
			edge:     nil,
			fromPort: 0,
			toPort:   0,
			assertSetup: func(t *testing.T, link *Link, err error) {
				assert.NotNil(t, link)
				assert.EqualError(t, err, fmt.Sprintf(util.ErrNilArgument, "edge"))
			},
		},
		"Valid link [Success]": {
			edge:     simpleEdge,
			fromPort: -1,
			toPort:   10,
			assertSetup: func(t *testing.T, link *Link, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, link)
				assert.EqualValues(t, simpleEdge, link.topoEdge)
				assert.Equal(t, int64(-1), link.fromPort)
				assert.Equal(t, int64(10), link.toPort)
			},
		},
		"Getters [Success]": {
			edge:     simpleEdge,
			fromPort: 4,
			toPort:   -8,
			assertSetup: func(t *testing.T, link *Link, err error) {
				assert.NotNil(t, link.TopoEdge())
				assert.NotNil(t, link.FromPort())
				assert.NotNil(t, link.ToPort())
				assert.EqualValues(t, simpleEdge, link.TopoEdge())
				assert.Equal(t, int64(4), link.fromPort)
				assert.Equal(t, int64(-8), link.toPort)
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			link, err := NewLink(tc.edge, tc.fromPort, tc.toPort)
			// Assert the result
			if tc.assertSetup != nil {
				tc.assertSetup(t, link, err)
			}
		})
	}
}
