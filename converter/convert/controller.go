package convert

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var controllerId int64

func init() {
	controllerId = 0
}

type Controller struct {
	id            int64
	switches      []*Switch
	newFlowTables om.OrderedMap[int64, *FlowTable]
}

func (c *Controller) ID() int64 {
	return c.id
}

func (c *Controller) Switches() []*Switch {
	return c.switches
}

func (c *Controller) NewFlowTables() *om.OrderedMap[int64, *FlowTable] {
	return &c.newFlowTables
}

func NewController(switches []*Switch) (*Controller, error) {
	if err := validateSwitches(switches); err != nil {
		return &Controller{}, err
	}

	c := &Controller{
		id:            controllerId,
		switches:      switches,
		newFlowTables: *om.New[int64, *FlowTable](),
	}
	controllerId++

	for _, s := range switches {
		s.SetController(c)
	}

	return c, nil
}

func validateSwitches(switches []*Switch) error {
	for _, s := range switches {
		if s == nil {
			return util.NewError(util.ErrNilInArray, "switches")
		}
	}

	return nil
}

func (c *Controller) FindSwitch(nodeId int64) *Switch {
	for _, sw := range c.switches {
		if sw.topoNode.ID() == nodeId {
			return sw
		}
	}
	return nil
}

/*
Adds flow rules to the new flow table of the switch with the given node id, creating the
new flow table if it doesn't exist. If the duplicateSwFT flag is set, the new flow table
is a copy of the switch's current flow table and is created only if new flow rules exist
in the given array.
*/
func (c *Controller) AddNewFlowRules(
	nodeId, destHostId int64,
	frs []FlowRule,
	duplicateSwFT bool,
) error {
	sw := c.FindSwitch(nodeId)
	if sw == nil {
		return util.NewError(util.ErrNoSwitchWithNodeId, nodeId)
	}

	ft, exists := c.newFlowTables.Get(nodeId)
	if !exists {
		if duplicateSwFT && !newEntriesExist(sw.FlowTable(), destHostId, frs) {
			return nil
		}
		ft = c.CreateNewFlowTable(sw, duplicateSwFT)
	}

	for _, fr := range frs {
		ft.AddEntry(destHostId, fr)
	}

	return nil
}

func (c *Controller) CreateNewFlowTable(
	sw *Switch,
	duplicateSwFT bool,
) *FlowTable {
	if duplicateSwFT {
		c.newFlowTables.Set(sw.topoNode.ID(), sw.FlowTable().Copy())
	} else {
		c.newFlowTables.Set(sw.topoNode.ID(), NewFlowTable())
	}
	ft, _ := c.newFlowTables.Get(sw.topoNode.ID())
	return ft
}

func newEntriesExist(
	swFt *FlowTable,
	destHostId int64,
	frs []FlowRule,
) bool {
	if swFt == nil {
		return true
	}

	for _, fr := range frs {
		if !swFt.hasEntry(destHostId, fr) {
			return true
		}
	}
	return false
}
