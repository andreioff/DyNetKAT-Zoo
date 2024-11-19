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
		flowTable:  nil,
		links:      []*Link{},
	}

	cases := map[string]struct {
		switchPort  int64
		sw          *Switch
		assertSetup func(*testing.T, int64, *Host, error)
	}{
		"Nil switch [Validation error]": {
			switchPort: 0,
			sw:         nil,
			assertSetup: func(t *testing.T, nextHostId int64, host *Host, err error) {
				assert.NotNil(t, host)
				assert.EqualError(t, err, fmt.Sprintf(util.ErrNilArgument, "sw"))
			},
		},
		"Valid Host [Success]": {
			switchPort: 4,
			sw:         mockSw,
			assertSetup: func(t *testing.T, nextHostId int64, host *Host, err error) {
				assert.Nil(t, err)
				assert.NotNil(t, host)

				assert.Greater(t, nextHostId, host.id)

				assert.Equal(t, int64(4), host.switchPort)

				assert.NotNil(t, host.sw)
				assert.Equal(t, int64(0), host.sw.topoNode.ID())
			},
		},
		"Host Getters [Success]": {
			switchPort: -3,
			sw:         mockSw,
			assertSetup: func(t *testing.T, nextHostId int64, host *Host, err error) {
				assert.Greater(t, nextHostId, host.ID())

				assert.Equal(t, int64(-3), host.SwitchPort())

				assert.NotNil(t, host.Switch())
				assert.Equal(t, int64(0), host.Switch().topoNode.ID())
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			host, err := NewHost(tc.switchPort, tc.sw)
			// Assert the result
			if tc.assertSetup != nil {
				tc.assertSetup(t, hostId, host, err)
			}
		})
	}
}
