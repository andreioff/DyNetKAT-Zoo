package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var (
	mockEdge1 graph.Edge  = simple.Edge{F: simple.Node(0), T: simple.Node(1)}
	mockEdge2 graph.Edge  = simple.Edge{F: simple.Node(1), T: simple.Node(3)}
	mockLink1 *Link       = &Link{topoEdge: mockEdge1, fromPort: 10, toPort: 11}
	mockLink2 *Link       = &Link{topoEdge: mockEdge2, fromPort: 21, toPort: 22}
	mockC     *Controller = &Controller{
		id:            1,
		switches:      []*Switch{},
		newFlowTables: make(map[int64]*FlowTable),
	}
)

func TestNewSwitch(t *testing.T) {
	type args struct {
		node  graph.Node
		links []*Link
	}
	tests := []struct {
		name    string
		args    args
		want    *Switch
		wantErr string
	}{
		{
			name: "Nil node [Validation error]",
			args: args{
				node:  nil,
				links: []*Link{},
			},
			want:    &Switch{},
			wantErr: fmt.Sprintf(util.ErrNilArgument, "node"),
		},
		{
			name: "Switch with non-incident link [Validation error]",
			args: args{
				node: simple.Node(1),
				links: []*Link{
					{ // incident
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{ // not incident
						topoEdge: simple.Edge{F: simple.Node(5), T: simple.Node(2)},
						fromPort: 9,
						toPort:   10,
					},
				},
			},
			want:    &Switch{},
			wantErr: fmt.Sprintf(util.ErrOnlyIncidentLinksForSwitch),
		},
		{
			name: "Valid switch no links [Sucess]",
			args: args{
				node:  simple.Node(1),
				links: []*Link{},
			},
			want: &Switch{
				topoNode:   simple.Node(1),
				links:      []*Link{},
				controller: nil,
				flowTable:  getMockEmptyFT(),
			},
			wantErr: "",
		},
		{
			name: "Valid switch with links [Sucess]",
			args: args{
				node:  simple.Node(1),
				links: []*Link{mockLink1, mockLink2},
			},
			want: &Switch{
				topoNode:   simple.Node(1),
				links:      []*Link{mockLink1, mockLink2},
				controller: nil,
				flowTable:  getMockEmptyFT(),
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewSwitch(tt.args.node, tt.args.links)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestSwitchGetters(t *testing.T) {
	type fields struct {
		topoNode   graph.Node
		controller *Controller
		flowTable  *FlowTable
		links      []*Link
	}
	tests := []struct {
		name        string
		fields      fields
		assertSetup func(*testing.T, *Switch)
	}{
		{
			name: "Switch getters [Sucess]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: mockC,
				flowTable:  getMockEmptyFT(),
				links:      []*Link{mockLink1, mockLink2},
			},
			assertSetup: func(t *testing.T, sw *Switch) {
				assert.NotNil(t, sw)

				assert.NotNil(t, sw.Controller())
				assert.EqualValues(t, mockC, sw.Controller())

				assert.NotNil(t, sw.TopoNode())
				assert.EqualValues(t, simple.Node(1), sw.TopoNode())

				assert.NotNil(t, sw.FlowTable())
				assert.EqualValues(t, getMockEmptyFT(), sw.FlowTable())
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sw := &Switch{
				topoNode:   tt.fields.topoNode,
				controller: tt.fields.controller,
				flowTable:  tt.fields.flowTable,
				links:      tt.fields.links,
			}
			// Assert the result
			if tt.assertSetup != nil {
				tt.assertSetup(t, sw)
			}
		})
	}
}

func TestSwitch_SetController(t *testing.T) {
	type fields struct {
		topoNode   graph.Node
		controller *Controller
		flowTable  *FlowTable
		links      []*Link
	}
	type args struct {
		c *Controller
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Controller
	}{
		{
			name: "Set valid controller [Success]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  nil,
				links:      []*Link{},
			},
			args: args{
				c: mockC,
			},
			want: mockC,
		},
		{
			name: "Set nil controller [Success]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: mockC,
				flowTable:  nil,
				links:      []*Link{},
			},
			args: args{
				c: nil,
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Switch{
				topoNode:   tt.fields.topoNode,
				controller: tt.fields.controller,
				flowTable:  tt.fields.flowTable,
				links:      tt.fields.links,
			}
			s.SetController(tt.args.c)
			assert.EqualValues(t, tt.want, s.controller)
		})
	}
}

func TestSwitch_GetLinkPorts(t *testing.T) {
	type fields struct {
		topoNode   graph.Node
		controller *Controller
		flowTable  *FlowTable
		links      []*Link
	}
	type args struct {
		otherNodeId int64
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantFromPort int64
		wantToPort   int64
		wantErr      string
	}{
		{
			name: "Empty link array [Error]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  getMockEmptyFT(),
				links:      []*Link{},
			},
			args:    args{-1},
			wantErr: util.ErrNoLinkBetweenSwitches,
		},
		{
			name: "No link found [Error]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  getMockEmptyFT(),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(5), T: simple.Node(1)},
						fromPort: 9,
						toPort:   10,
					},
				},
			},
			args:    args{6},
			wantErr: util.ErrNoLinkBetweenSwitches,
		},
		{
			name: "Outgoing link [Success]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  getMockEmptyFT(),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
				},
			},
			args:         args{2},
			wantFromPort: 3,
			wantToPort:   4,
		},
		{
			name: "Incoming link [Success]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  getMockEmptyFT(),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
				},
			},
			args:         args{4},
			wantFromPort: 8,
			wantToPort:   7,
		},
		{
			name: "Duplicate link [Success]",
			fields: fields{
				topoNode:   simple.Node(1),
				controller: nil,
				flowTable:  getMockEmptyFT(),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(4)},
						fromPort: 9,
						toPort:   10,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
				},
			},
			args:         args{4},
			wantFromPort: 9,
			wantToPort:   10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Switch{
				topoNode:   tt.fields.topoNode,
				controller: tt.fields.controller,
				flowTable:  tt.fields.flowTable,
				links:      tt.fields.links,
			}
			gotFromPort, gotToPort, err := s.GetLinkPorts(tt.args.otherNodeId)
			if tt.wantErr == "" {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
			assert.Equal(t, tt.wantFromPort, gotFromPort)
			assert.Equal(t, tt.wantToPort, gotToPort)
		})
	}
}

func Test_validateLinks(t *testing.T) {
	type args struct {
		node  graph.Node
		links []*Link
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "Nil node [Error]",
			args: args{
				node:  nil,
				links: []*Link{},
			},
			wantErr: fmt.Sprintf(util.ErrNilArgument, "node"),
		},
		{
			name: "Empty links [Success]",
			args: args{
				node:  simple.Node(1),
				links: []*Link{},
			},
		},
		{
			name: "Only incident links [Success]",
			args: args{
				node: simple.Node(1),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(4), T: simple.Node(1)},
						fromPort: 7,
						toPort:   8,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(5), T: simple.Node(1)},
						fromPort: 9,
						toPort:   10,
					},
				},
			},
		},
		{
			name: "Non-incident link [Success]",
			args: args{
				node: simple.Node(1),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(3)},
						fromPort: 5,
						toPort:   6,
					},
					{
						topoEdge: simple.Edge{F: simple.Node(5), T: simple.Node(4)},
						fromPort: 9,
						toPort:   10,
					},
				},
			},
			wantErr: util.ErrOnlyIncidentLinksForSwitch,
		},
		{
			name: "Nil links [Success]",
			args: args{
				node: simple.Node(1),
				links: []*Link{
					{
						topoEdge: simple.Edge{F: simple.Node(1), T: simple.Node(2)},
						fromPort: 3,
						toPort:   4,
					},
					nil,
				},
			},
			wantErr: fmt.Sprintf(util.ErrNilInArray, "links"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotErr := validateLinks(tt.args.node, tt.args.links)
			if tt.wantErr == "" {
				assert.NoError(t, gotErr)
			} else {
				assert.Error(t, gotErr)
				assert.Equal(t, tt.wantErr, gotErr.Error())
			}
		})
	}
}
