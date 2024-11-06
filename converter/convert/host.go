package convert

type Host struct {
	id         int64
	switchPort int64
}

func NewHost(id, switchPort int64) *Host {
	return &Host{
		id:         id,
		switchPort: switchPort,
	}
}

func (h Host) ID() int64 {
	return h.id
}

func (h Host) SwitchPort() int64 {
	return h.switchPort
}
