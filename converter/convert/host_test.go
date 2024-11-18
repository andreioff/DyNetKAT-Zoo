package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/graph/simple"

	"utwente.nl/topology-to-dynetkat-coverter/util"
)

func TestConvert_NewHost(t *testing.T) {
	mockSw := &Switch{
		topoNode:   simple.Node(0),
		controller: nil,
		hosts:      []*Host{},
		flowTable:  nil,
		links:      []*Link{},
	}

	cases := map[string]struct {
		id          int64
		switchPort  int64
		sw          *Switch
		assertSetup func(*testing.T, Host, error)
	}{
		"Nil switch [Validation error]": {
			id:         0,
			switchPort: 0,
			sw:         nil,
			assertSetup: func(t *testing.T, host Host, err error) {
				assert.NotNil(t, host)
				assert.EqualError(t, err, fmt.Sprintf(util.ErrNilArgument, "sw"))
			},
		},
		"Valid Host [Success]": {
			id:         -1,
			switchPort: 4,
			sw:         mockSw,
			assertSetup: func(t *testing.T, host Host, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, host)
				assert.Equal(t, int64(-1), host.id)
				assert.Equal(t, int64(4), host.switchPort)
				assert.NotNil(t, host.sw)
				assert.Equal(t, int64(0), host.sw.TopoNode().ID())
			},
		},
		"Host Getters [Success]": {
			id:         1,
			switchPort: -3,
			sw:         mockSw,
			assertSetup: func(t *testing.T, host Host, err error) {
				assert.Equal(t, int64(1), host.ID())
				assert.Equal(t, int64(-3), host.SwitchPort())
				assert.NotNil(t, host.Switch())
				assert.Equal(t, int64(0), host.Switch().TopoNode().ID())
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			host, err := NewHost(tc.id, tc.switchPort, tc.sw)
			// Assert the result
			if tc.assertSetup != nil {
				tc.assertSetup(t, host, err)
			}
		})
	}
}
