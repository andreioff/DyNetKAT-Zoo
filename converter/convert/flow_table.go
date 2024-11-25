package convert

import (
	"strconv"
)

const (
	DST_STRING  = "dst"
	PORT_STRING = "port"
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

/*
Returns a pointer to a new flow table containing only
the flow rules of the current flow table that satisfy
the given predicate.
*/
func (ft *FlowTable) Filter(pred func(FlowRule) bool) *FlowTable {
	entries := make(map[int64][]FlowRule)

	for hostId, frs := range ft.entries {
		newFrs := []FlowRule{}
		for _, fr := range frs {
			if pred(fr) {
				newFrs = append(newFrs, fr)
			}
		}

		if len(newFrs) > 0 {
			entries[hostId] = newFrs
		}
	}

	newFt := NewFlowTable()
	newFt.setEntries(entries)
	return newFt
}

// Extends the current flow table with the entries of the
// given flow table
func (ft *FlowTable) Extend(otherFt *FlowTable) {
	if otherFt == nil {
		return
	}

	for hostId, frs := range otherFt.entries {
		for _, fr := range frs {
			ft.AddEntry(hostId, fr)
		}
	}
}

func (ft *FlowTable) ToNetKATPolicies() []*SimpleNetKATPolicy {
	policies := []*SimpleNetKATPolicy{}

	for destHostId, frs := range ft.entries {
		for _, fr := range frs {
			policy := NewSimpleNetKATPolicy()
			policy.AddTest(DST_STRING, strconv.FormatInt(destHostId, 10))
			policy.AddTest(PORT_STRING, strconv.FormatInt(fr.inPort, 10))
			policy.AddAssignment(PORT_STRING, strconv.FormatInt(fr.outPort, 10))
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
