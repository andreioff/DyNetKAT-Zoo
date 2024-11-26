package convert

import (
	"strconv"

	om "github.com/wk8/go-ordered-map/v2"
)

const (
	DST_STRING  = "dst"
	PORT_STRING = "port"
)

type (
	ftKeyT = int64
	ftValT = []FlowRule
)

type FlowTable struct {
	entries om.OrderedMap[ftKeyT, ftValT] // maps host destination id to corresponding flow rules
}

func (ft *FlowTable) Entries() *om.OrderedMap[ftKeyT, ftValT] {
	return &ft.entries
}

func (ft *FlowTable) setEntries(newEntries om.OrderedMap[ftKeyT, ftValT]) {
	ft.entries = newEntries
}

func NewFlowTable() *FlowTable {
	return &FlowTable{
		entries: *om.New[ftKeyT, ftValT](),
	}
}

func (ft *FlowTable) AddEntry(destHostId int64, fr FlowRule) {
	// do not add duplicate entries
	if ft.hasEntry(destHostId, fr) {
		return
	}

	destFrsPair := ft.entries.GetPair(destHostId)
	if destFrsPair == nil {
		ft.entries.Set(destHostId, []FlowRule{fr})
		return
	}

	destFrsPair.Value = append(destFrsPair.Value, fr)
}

func (ft *FlowTable) hasEntry(hostId int64, target FlowRule) bool {
	frs, exists := ft.entries.Get(hostId)
	if !exists {
		return false
	}

	for _, fr := range frs {
		if fr.IsEqual(target) {
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
	entries := *om.New[ftKeyT, ftValT]()

	for pair := ft.entries.Oldest(); pair != nil; pair = pair.Next() {
		hostId, frs := pair.Key, pair.Value
		newFrs := []FlowRule{}
		for _, fr := range frs {
			if pred(fr) {
				newFrs = append(newFrs, fr)
			}
		}

		if len(newFrs) > 0 {
			entries.Set(hostId, newFrs)
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

	for pair := otherFt.entries.Oldest(); pair != nil; pair = pair.Next() {
		hostId, frs := pair.Key, pair.Value
		for _, fr := range frs {
			ft.AddEntry(hostId, fr)
		}
	}
}

func (ft *FlowTable) ToNetKATPolicies() []*SimpleNetKATPolicy {
	policies := []*SimpleNetKATPolicy{}

	for pair := ft.entries.Oldest(); pair != nil; pair = pair.Next() {
		destHostId, frs := pair.Key, pair.Value
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
	entries := *om.New[ftKeyT, ftValT]()

	for pair := ft.entries.Oldest(); pair != nil; pair = pair.Next() {
		hostId, frs := pair.Key, pair.Value
		newFrs := make([]FlowRule, len(frs))
		copy(newFrs, frs)

		entries.Set(hostId, newFrs)
	}
	newFt.setEntries(entries)

	return newFt
}
