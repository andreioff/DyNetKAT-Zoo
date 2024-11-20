package convert

import (
	"strconv"
)

type FlowTable struct {
	entries map[int64][]FlowRule // maps host destination id to corresponding flow rules
}

func (ft *FlowTable) Entries() map[int64][]FlowRule {
	return ft.entries
}

func (ft *FlowTable) setEntries(newEntries map[int64][]FlowRule) {
	ft.entries = newEntries
}

func NewFlowTable() *FlowTable {
	return &FlowTable{
		entries: make(map[int64][]FlowRule),
	}
}

func (ft *FlowTable) AddEntry(destHostId int64, fr FlowRule) {
	// do not add duplicate entries
	if ft.hasEntry(destHostId, fr) {
		return
	}

	ft.entries[destHostId] = append(ft.entries[destHostId], fr)
}

func (ft *FlowTable) hasEntry(hostId int64, fr FlowRule) bool {
	if _, exists := ft.entries[hostId]; !exists {
		return false
	}

	for _, v := range ft.entries[hostId] {
		if v.IsEqual(fr) {
			return true
		}
	}

	return false
}

func (ft *FlowTable) ToNetKATPolicies() []*SimpleNetKATPolicy {
	policies := []*SimpleNetKATPolicy{}

	for destHostId, frs := range ft.entries {
		for _, fr := range frs {
			policy := NewSimpleNetKATPolicy()
			policy.AddTest("dst", strconv.FormatInt(destHostId, 10))
			policy.AddTest("port", strconv.FormatInt(fr.inPort, 10))
			policy.AddAssignment("port", strconv.FormatInt(fr.outPort, 10))
			policies = append(policies, policy)
		}
	}

	return policies
}

// returns a deep copy of this flow table
func (ft *FlowTable) Copy() *FlowTable {
	newFt := NewFlowTable()
	entries := make(map[int64][]FlowRule)

	for hostId, frs := range ft.entries {
		newFrs := make([]FlowRule, len(frs))
		copy(newFrs, frs)

		entries[hostId] = newFrs
	}
	newFt.setEntries(entries)

	return newFt
}
