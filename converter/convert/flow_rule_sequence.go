package convert

type FRSequenceEntry struct {
	SwitchId   int64
	DestHostId int64
	FlowRule   FlowRule
}

func NewFrSequenceEntry(switchId int64, destHostId int64, fr FlowRule) FRSequenceEntry {
	return FRSequenceEntry{
		switchId,
		destHostId,
		fr,
	}
}

func (e1 FRSequenceEntry) isEqual(e2 FRSequenceEntry) bool {
	return e1.SwitchId == e2.SwitchId &&
		e1.DestHostId == e2.DestHostId &&
		e1.FlowRule.IsEqual(e2.FlowRule)
}

type FlowRuleSequence struct {
	// a sequence of flow rules to be sent in order by a
	// controller to the corresponding switches
	entries []FRSequenceEntry
}

func NewFlowRuleSequence() *FlowRuleSequence {
	return &FlowRuleSequence{
		entries: []FRSequenceEntry{},
	}
}

func (frs *FlowRuleSequence) Entries() []FRSequenceEntry {
	return frs.entries
}

func (frs *FlowRuleSequence) AddEntry(switchId int64, destHostId int64, fr FlowRule) {
	newEntry := NewFrSequenceEntry(switchId, destHostId, fr)
	// do not add duplicate entries
	if frs.hasEntry(newEntry) {
		return
	}

	frs.entries = append(frs.entries, newEntry)
}

func (frs *FlowRuleSequence) hasEntry(target FRSequenceEntry) bool {
	for _, e := range frs.entries {
		if e.isEqual(target) {
			return true
		}
	}

	return false
}

/*
Returns a pointer to a new flow rule sequence containing only
the flow rules of the given sequence that satisfy
the given predicate.
*/
func (frs *FlowRuleSequence) Filter(pred func(FRSequenceEntry) bool) *FlowRuleSequence {
	filtered := []FRSequenceEntry{}
	for _, e := range frs.entries {
		if pred(e) {
			filtered = append(filtered, e)
		}
	}

	return &FlowRuleSequence{filtered}
}
