package encode

import (
	"fmt"

	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type ControllerUpdate struct {
	FlowTables  om.OrderedMap[int64, *convert.FlowTable]
	FrSequences []*convert.FlowRuleSequence
}

type EncodingInfo struct {
	NodeIdToIndex   om.OrderedMap[int64, int]                // maps switch node id to index
	UsedSwitchFTs   om.OrderedMap[int64, *convert.FlowTable] // maps switch node id to flow table of switch
	UsedContUpdates []ControllerUpdate                       // maps switch node id to new flow table
	Links           []convert.SimpleNetKATPolicy
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedContUpdates := getUsedControllerUpdates(n.Controllers())
	links := getLinksAsNetKATPolicies(n.Controllers(), n.Switches())

	if usedSwitchFTs.Len() == 0 || len(usedContUpdates) == 0 {
		return EncodingInfo{}, util.NewError(util.ErrNoSwsOrContsUsed)
	}

	return EncodingInfo{
		NodeIdToIndex:   *om.New[int64, int](),
		UsedSwitchFTs:   usedSwitchFTs,
		UsedContUpdates: usedContUpdates,
		Links:           links,
	}, nil
}

func getUsedSwitchesFTs(switches []*convert.Switch) om.OrderedMap[int64, *convert.FlowTable] {
	usedSwitchFTs := *om.New[int64, *convert.FlowTable]()

	for _, sw := range switches {
		filteredFT := sw.FlowTable().Filter(func(fr convert.FlowRule) bool {
			return !fr.IsLink()
		})

		isUpdatingSwitch := false
		for _, c := range sw.Controllers() {
			if c != nil && c.IsUpdatingSwitch(sw.TopoNode().ID()) {
				isUpdatingSwitch = true
			}
		}
		if filteredFT.Entries().Len() > 0 || isUpdatingSwitch {
			usedSwitchFTs.Set(sw.TopoNode().ID(), filteredFT)
		}
	}

	return usedSwitchFTs
}

func getUsedControllerUpdates(
	controllers []*convert.Controller,
) []ControllerUpdate {
	usedControllerUpdates := []ControllerUpdate{}
	ftPred := func(fr convert.FlowRule) bool { return !fr.IsLink() }
	frsPred := func(e convert.FRSequenceEntry) bool { return !e.FlowRule.IsLink() }

	for _, c := range controllers {
		cUpdate := ControllerUpdate{
			FlowTables:  *om.New[int64, *convert.FlowTable](),
			FrSequences: []*convert.FlowRuleSequence{},
		}
		for pair := c.NewFlowTables().Oldest(); pair != nil; pair = pair.Next() {
			cUpdate.FlowTables.Set(pair.Key, pair.Value.Filter(ftPred))
		}
		for _, frs := range c.NewFlowRuleSequences() {
			cUpdate.FrSequences = append(cUpdate.FrSequences, frs.Filter(frsPred))
		}

		if cUpdate.FlowTables.Len() > 0 || len(cUpdate.FrSequences) > 0 {
			usedControllerUpdates = append(usedControllerUpdates, cUpdate)
		}
	}

	return usedControllerUpdates
}

func getLinksAsNetKATPolicies(
	controllers []*convert.Controller,
	switches []*convert.Switch,
) []convert.SimpleNetKATPolicy {
	ftPred := func(fr convert.FlowRule) bool { return fr.IsLink() }
	frsPred := func(e convert.FRSequenceEntry) bool { return e.FlowRule.IsLink() }

	links := []convert.SimpleNetKATPolicy{}
	for _, c := range controllers {
		for pair := c.NewFlowTables().Oldest(); pair != nil; pair = pair.Next() {
			links = append(links, pair.Value.Filter(ftPred).ToNetKATPolicies()...)
		}
		for _, frs := range c.NewFlowRuleSequences() {
			for _, e := range frs.Filter(frsPred).Entries() {
				links = append(links, e.ToNetKATPolicy())
			}
		}
	}

	for _, sw := range switches {
		links = append(links, sw.FlowTable().Filter(ftPred).ToNetKATPolicies()...)
	}
	return links
}

func (ei EncodingInfo) FindNewFT(nodeId int64) (*convert.FlowTable, bool) {
	for _, update := range ei.UsedContUpdates {
		newFt, exists := update.FlowTables.Get(nodeId)
		if exists {
			return newFt, true
		}
	}

	return nil, false
}

func (ei EncodingInfo) GetSwIndex(nodeId int64) int {
	indx, exists := ei.NodeIdToIndex.Get(nodeId)
	if !exists {
		ei.NodeIdToIndex.Set(nodeId, ei.NodeIdToIndex.Len())
		indx, _ = ei.NodeIdToIndex.Get(nodeId)
	}

	return indx
}

func (ei EncodingInfo) GetSwIndexStr(nodeId int64) string {
	indx := ei.GetSwIndex(nodeId)
	return fmt.Sprintf("%d", indx)
}

/*
Returns a deep copy of the flow table of every switch
*/
func (ei EncodingInfo) CopyUsedSwitchesFTs() *om.OrderedMap[int64, *convert.FlowTable] {
	flowTableCopy := om.New[int64, *convert.FlowTable]()
	for pair := ei.UsedSwitchFTs.Oldest(); pair != nil; pair = pair.Next() {
		flowTableCopy.Set(pair.Key, pair.Value.Copy())
	}
	return flowTableCopy
}
