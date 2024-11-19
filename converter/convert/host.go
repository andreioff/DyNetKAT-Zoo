package convert

import (
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

var hostId int64

func init() {
	hostId = 0
}

type Host struct {
	id         int64
	switchPort int64
	sw         *Switch
}

func NewHost(switchPort int64, sw *Switch) (*Host, error) {
	if sw == nil {
		return &Host{}, util.NewError(util.ErrNilArgument, "sw")
	}

	host := &Host{
		id:         hostId,
		switchPort: switchPort,
		sw:         sw,
	}
	hostId++

	return host, nil
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
