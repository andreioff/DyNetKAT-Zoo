package convert

import "errors"

type Host struct {
	id         int64
	switchPort int64
	sw         *Switch
}

func NewHost(id, switchPort int64, sw *Switch) (Host, error) {
	if sw == nil {
		return Host{}, errors.New("Received nil switch!")
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
