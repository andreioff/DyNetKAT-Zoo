package convert

import (
	"strconv"

	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type FlowTable struct {
	entries map[util.I64Tup][]int64 // maps host destination id and incoming port to outgoing port
}

func (ft *FlowTable) Entries() map[util.I64Tup][]int64 {
	return ft.entries
}

func (ft *FlowTable) setEntries(newEntries map[util.I64Tup][]int64) {
	ft.entries = newEntries
}

func NewFlowTable() *FlowTable {
	return &FlowTable{
		entries: make(map[util.I64Tup][]int64),
	}
}

func (ft *FlowTable) AddEntry(destHostId, inPort, outPort int64) {
	key := util.NewI64Tup(destHostId, inPort)

	// do not add duplicate entries
	if ft.hasEntry(key, outPort) {
		return
	}

	ft.entries[key] = append(ft.entries[key], outPort)
}

func (ft *FlowTable) hasEntry(key util.I64Tup, value int64) bool {
	if _, exists := ft.entries[key]; !exists {
		return false
	}

	for _, v := range ft.entries[key] {
		if v == value {
			return true
		}
	}

	return false
}

func (ft *FlowTable) ToNetKATPolicies() []*SimpleNetKATPolicy {
	policies := []*SimpleNetKATPolicy{}

	for hostIdInPort, outPorts := range ft.entries {
		dstHostId, inPort := hostIdInPort.Fst, hostIdInPort.Snd
		for _, outPort := range outPorts {
			policy := NewSimpleNetKATPolicy()
			policy.AddTest("dst", strconv.FormatInt(dstHostId, 10))
			policy.AddTest("port", strconv.FormatInt(inPort, 10))
			policy.AddAssignment("port", strconv.FormatInt(outPort, 10))
			policies = append(policies, policy)
		}
	}

	return policies
}

// returns a deep copy of this flow table
func (ft *FlowTable) Copy() *FlowTable {
	newFt := NewFlowTable()
	entries := make(map[util.I64Tup][]int64)

	for hostIdInPort, outPorts := range ft.entries {
		newOutPorts := make([]int64, len(outPorts))
		copy(newOutPorts, outPorts)

		entries[hostIdInPort] = newOutPorts
	}
	newFt.setEntries(entries)

	return newFt
}
