package convert

import (
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Host struct {
	id         int64
	switchPort int64
	sw         *Switch
}

func NewHost(id, switchPort int64, sw *Switch) (Host, error) {
	if sw == nil {
		return Host{}, util.NewError(util.ErrNilArgument, "sw")
	}

	return Host{
		id:         id,
		switchPort: switchPort,
		sw:         sw,
	}, nil
}

func (h *Host) ID() int64 {
	return h.id
}

func (h *Host) SwitchPort() int64 {
	return h.switchPort
}

func (h *Host) Switch() *Switch {
	return h.sw
}
