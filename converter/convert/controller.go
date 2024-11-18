package convert

import (
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var id int64

func init() {
	id = 0
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

func NewController(switches []*Switch) *Controller {
	c := &Controller{
		id:            id,
		switches:      switches,
		newFlowTables: make(map[int64]*FlowTable),
	}
	id++

	for _, s := range switches {
		s.SetController(c)
	}

	return c
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
		if !c.newEntriesExist(sw.FlowTable(), destHostId, portTups) {
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

func (c *Controller) newEntriesExist(
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
