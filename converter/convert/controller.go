package convert

import (
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var controllerId int64

func init() {
	controllerId = 0
}

type Controller struct {
	id            int64
	switches      []*Switch
	newFlowTables map[int64]*FlowTable
}

func (c *Controller) ID() int64 {
	return c.id
}

func (c *Controller) Switches() []*Switch {
	return c.switches
}

func (c *Controller) NewFlowTables() map[int64]*FlowTable {
	return c.newFlowTables
}

func NewController(switches []*Switch) (*Controller, error) {
	if err := validateSwitches(switches); err != nil {
		return &Controller{}, err
	}

	c := &Controller{
		id:            controllerId,
		switches:      switches,
		newFlowTables: make(map[int64]*FlowTable),
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

func (c *Controller) findSwitch(nodeId int64) *Switch {
	for _, sw := range c.switches {
		if sw.topoNode.ID() == nodeId {
			return sw
		}
	}
	return nil
}

/*
Adds flow rules to the new flow table of the switch with the given node id, creating the
new flow table if it doesn't exist. The flow table is created only if new flow rules exist.
*/
func (c *Controller) AddNewFlowRules(nodeId, destHostId int64, portTups []util.I64Tup) error {
	sw := c.findSwitch(nodeId)
	if sw == nil {
		return util.NewError(util.ErrNoSwitchWithNodeId, nodeId)
	}

	ft, exists := c.newFlowTables[nodeId]
	if !exists {
		if !newEntriesExist(sw.FlowTable(), destHostId, portTups) {
			return nil
		}
		c.newFlowTables[nodeId] = sw.FlowTable().Copy()
		ft = c.newFlowTables[nodeId]
	}

	for _, inPortOutPort := range portTups {
		ft.AddEntry(destHostId, inPortOutPort.Fst, inPortOutPort.Snd)
	}

	return nil
}

func newEntriesExist(
	swFt *FlowTable,
	destHostId int64,
	portTups []util.I64Tup,
) bool {
	if swFt == nil {
		return true
	}

	for _, inPortOutPort := range portTups {
		hasEntry := swFt.hasEntry(util.NewI64Tup(destHostId, inPortOutPort.Fst), inPortOutPort.Snd)
		if !hasEntry {
			return true
		}
	}
	return false
}
