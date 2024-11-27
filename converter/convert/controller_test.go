package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	om "github.com/wk8/go-ordered-map/v2"
	"gonum.org/v1/gonum/graph/simple"

	tu "utwente.nl/topology-to-dynetkat-coverter/test_util"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

func TestNewController(t *testing.T) {
	type K = int64
	type V = *FlowTable

	type args struct {
		switches []*Switch
	}
	tests := []struct {
		name        string
		args        args
		wantErr     string
		assertSetup func(*testing.T, int64, *Controller)
	}{
		{
			name: "No switches [Success]",
			args: args{
				switches: []*Switch{},
			},
			assertSetup: func(t *testing.T, nextContId int64, c *Controller) {
				assert.NotNil(t, c)

				assert.Greater(t, nextContId, c.id)

				assert.ElementsMatch(t, []*Switch{}, c.switches)
				tu.AssertEqualMaps(t, om.New[K, V](), &c.newFlowTables)
			},
		},
		{
			name: "With switches [Success]",
			args: args{
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  nil,
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  nil,
					},
				},
			},
			assertSetup: func(t *testing.T, nextContId int64, c *Controller) {
				assert.NotNil(t, c)

				assert.Greater(t, nextContId, c.id)

				tu.AssertEqualMaps(t, om.New[K, V](), &c.newFlowTables)
				assert.ElementsMatch(t, []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: c,
						flowTable:  nil,
					},
					{
						topoNode:   simple.Node(2),
						controller: c,
						flowTable:  nil,
					},
				}, c.switches)
			},
		},
		{
			name: "Getters [Success]",
			args: args{
				switches: []*Switch{
					{
						topoNode:   simple.Node(3),
						controller: nil,
						flowTable:  nil,
					},
				},
			},
			assertSetup: func(t *testing.T, nextContId int64, c *Controller) {
				assert.NotNil(t, c)

				assert.Greater(t, nextContId, c.ID())

				tu.AssertEqualMaps(t, om.New[K, V](), c.NewFlowTables())
				assert.ElementsMatch(t, []*Switch{
					{
						topoNode:   simple.Node(3),
						controller: c,
						flowTable:  nil,
					},
				}, c.Switches())
			},
		},
		{
			name: "Nil switches [Error]",
			args: args{
				switches: []*Switch{
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  nil,
					},
					nil,
				},
			},
			wantErr: fmt.Sprintf(util.ErrNilInArray, "switches"),
			assertSetup: func(t *testing.T, nextContId int64, c *Controller) {
				assert.NotNil(t, c)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewController(tt.args.switches)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
			// Assert the result
			if tt.assertSetup != nil {
				tt.assertSetup(t, controllerId, got)
			}
		})
	}
}

func TestController_FindSwitch(t *testing.T) {
	type K = int64
	type V = *FlowTable

	type fields struct {
		id            int64
		switches      []*Switch
		newFlowTables om.OrderedMap[K, V]
	}
	type args struct {
		nodeId int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Switch
	}{
		{
			name: "Empty switch array [Success]",
			fields: fields{
				id:            1,
				switches:      []*Switch{},
				newFlowTables: *om.New[K, V](),
			},
			args: args{
				nodeId: -1,
			},
			want: nil,
		},
		{
			name: "Non-empty array with no match [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  &FlowTable{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
					},
				},
				newFlowTables: *om.New[K, V](),
			},
			args: args{
				nodeId: 3,
			},
			want: nil,
		},
		{
			name: "Non-empty array with match [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  &FlowTable{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
					},
					{
						topoNode:   simple.Node(14),
						controller: nil,
						flowTable:  &FlowTable{},
					},
				},
				newFlowTables: *om.New[K, V](),
			},
			args: args{
				nodeId: 14,
			},
			want: &Switch{
				topoNode:   simple.Node(14),
				controller: nil,
				flowTable:  &FlowTable{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Controller{
				id:            tt.fields.id,
				switches:      tt.fields.switches,
				newFlowTables: tt.fields.newFlowTables,
			}
			got := c.FindSwitch(tt.args.nodeId)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func Test_newEntriesExist(t *testing.T) {
	type args struct {
		swFt       *FlowTable
		destHostId int64
		flowRules  []FlowRule
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Nil flow table [Success]",
			args: args{
				swFt:       nil,
				destHostId: 0,
				flowRules:  []FlowRule{},
			},
			want: true,
		},
		{
			name: "Empty entries [Success]",
			args: args{
				swFt:       getMockEmptyFT(),
				destHostId: 0,
				flowRules:  []FlowRule{},
			},
			want: false,
		},
		{
			name: "Existing entries [Success]",
			args: args{
				swFt:       getMockFT1(),
				destHostId: 3,
				flowRules: []FlowRule{
					{15, 16, false},
					{15, 17, true},
				},
			},
			want: false,
		},
		{
			name: "New entries [Success]",
			args: args{
				swFt:       getMockFT1(),
				destHostId: 4,
				flowRules: []FlowRule{
					{18, 19, false},
					{19, 20, true},
				},
			},
			want: true,
		},
		{
			name: "Mixed entries [Success]",
			args: args{
				swFt:       getMockFT1(),
				destHostId: 3,
				flowRules: []FlowRule{
					{15, 16, false},
					{15, 17, true},
					{18, 19, false},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newEntriesExist(tt.args.swFt, tt.args.destHostId, tt.args.flowRules)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestController_AddNewFlowRules(t *testing.T) {
	type K = int64
	type V = *FlowTable
	pair := func(key K, value V) om.Pair[K, V] {
		return om.Pair[K, V]{
			Key:   key,
			Value: value,
		}
	}
	newMap := func(data []om.Pair[K, V]) *om.OrderedMap[K, V] {
		return om.New[K, V](om.WithInitialData(data...))
	}

	type fields struct {
		id            int64
		switches      []*Switch
		newFlowTables om.OrderedMap[K, V]
	}
	type args struct {
		nodeId     int64
		destHostId int64
		flowRules  []FlowRule
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     string
		assertSetup func(*testing.T, *Controller, *Controller)
	}{
		{
			name: "No switch found [Error]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
				},
				newFlowTables: *om.New[K, V](),
			},
			args: args{
				nodeId:     -1,
				destHostId: 0,
				flowRules:  []FlowRule{},
			},
			wantErr: fmt.Sprintf(util.ErrNoSwitchWithNodeId, -1),
		},
		{
			name: "No new flow table, no new entries [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  getMockFT1(),
						links:      []*Link{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
				},
				newFlowTables: *newMap([]om.Pair[K, V]{
					pair(2, getMockFT2()),
				}),
			},
			args: args{
				nodeId:     1,
				destHostId: 1,
				flowRules:  []FlowRule{{13, 14, false}},
			},
			assertSetup: func(t *testing.T, c *Controller, initial *Controller) {
				assert.EqualValues(t, initial, c)
			},
		},
		{
			name: "No new flow table, new entries [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  getMockFT1(),
						links:      []*Link{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
				},
				newFlowTables: *newMap([]om.Pair[K, V]{
					pair(2, getMockFT2()),
				}),
			},
			args: args{
				nodeId:     1,
				destHostId: 1,
				flowRules:  []FlowRule{{13, 19, false}, {13, 20, true}},
			},
			assertSetup: func(t *testing.T, c *Controller, initial *Controller) {
				assert.ElementsMatch(t, initial.switches, c.switches)
				tu.AssertEqualMaps(
					t,
					newMap([]om.Pair[K, V]{
						pair(1, getMockFT3()),
						pair(2, getMockFT2()),
					}),
					&c.newFlowTables,
				)
			},
		},
		{
			name: "Existing new flow table, no new entries [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  getMockFT1(),
						links:      []*Link{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
				},
				newFlowTables: *newMap([]om.Pair[K, V]{
					pair(1, getMockFT3()),
					pair(2, getMockFT2()),
				}),
			},
			args: args{
				nodeId:     1,
				destHostId: 1,
				flowRules:  []FlowRule{{13, 19, false}, {13, 20, true}},
			},
			assertSetup: func(t *testing.T, c *Controller, initial *Controller) {
				assert.EqualValues(t, initial, c)
			},
		},
		{
			name: "Existing new flow table, new entries [Success]",
			fields: fields{
				id: 1,
				switches: []*Switch{
					{
						topoNode:   simple.Node(1),
						controller: nil,
						flowTable:  getMockFT1(),
						links:      []*Link{},
					},
					{
						topoNode:   simple.Node(2),
						controller: nil,
						flowTable:  &FlowTable{},
						links:      []*Link{},
					},
				},
				newFlowTables: *newMap([]om.Pair[K, V]{
					pair(1, getMockFT1()),
					pair(2, getMockFT2()),
				}),
			},
			args: args{
				nodeId:     1,
				destHostId: 3,
				flowRules:  []FlowRule{{15, 20, false}, {21, 22, true}},
			},
			assertSetup: func(t *testing.T, c *Controller, initial *Controller) {
				assert.ElementsMatch(t, initial.switches, c.switches)
				tu.AssertEqualMaps(t, newMap([]om.Pair[K, V]{
					pair(1, &FlowTable{
						*ftNewMap(
							ftPair(0, ftValT{
								{10, 11, false},
								{10, 12, true},
							}),
							ftPair(1, ftValT{{13, 14, false}}),
							ftPair(3, ftValT{
								{15, 16, false},
								{15, 17, true},
								{15, 20, false},
								{21, 22, true},
							}),
						),
					}),
					pair(2, getMockFT2()),
				}), &c.newFlowTables)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cInitial := &Controller{
				id:            tt.fields.id,
				switches:      tt.fields.switches,
				newFlowTables: tt.fields.newFlowTables,
			}
			c := *cInitial // copy

			err := c.AddNewFlowRules(tt.args.nodeId, tt.args.destHostId, tt.args.flowRules)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
				assert.EqualValues(t, &c, cInitial) // no side effects
			}
			// Assert the result
			if tt.assertSetup != nil {
				tt.assertSetup(t, &c, cInitial)
			}
		})
	}
}
