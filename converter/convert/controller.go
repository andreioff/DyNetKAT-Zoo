package convert

import (
	"errors"
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

/*
Copies the flow table of the switch with the given node id
to the 'newFlowTables' map. It overwrites the map entry if
the entry already exists!
Returns an error if the switch with the given node id is
not found, and nil otherwise.
*/
func (c *Controller) CopyFlowTable(nodeId int64) error {
	sw := c.findSwitch(nodeId)
	if sw == nil {
		return errors.New("No switch matches the given node id!")
	}

	c.newFlowTables[nodeId] = sw.FlowTable().Copy()

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
