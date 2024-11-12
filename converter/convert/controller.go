package convert

var id int64

func init() {
	id = 0
}

type Controller struct {
	id            int64
	switches      []*Switch
	newFlowTables []*Switch
}

func (c *Controller) ID() int64 {
	return c.id
}

func (c *Controller) Switches() []*Switch {
	return c.switches
}

func NewController(switches []*Switch) *Controller {
	c := &Controller{
		id:            id,
		switches:      switches,
		newFlowTables: []*Switch{},
	}
	id++

	for _, s := range switches {
		s.SetController(c)
	}

	return c
}
