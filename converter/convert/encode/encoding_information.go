package encode

import (
	"fmt"

	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type ControllerUpdate struct {
	flowTables  om.OrderedMap[int64, *convert.FlowTable]
	frSequences []*convert.FlowRuleSequence
}

type EncodingInfo struct {
	nodeIdToIndex   om.OrderedMap[int64, int]                // maps switch node id to index
	usedSwitchFTs   om.OrderedMap[int64, *convert.FlowTable] // maps switch node id to flow table of switch
	usedContUpdates []ControllerUpdate                       // maps switch node id to new flow table
	links           []convert.SimpleNetKATPolicy
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedContUpdates := getUsedControllerUpdates(n.Controllers())
	links := getLinksAsNetKATPolicies(n.Controllers(), n.Switches())

	if usedSwitchFTs.Len() == 0 || len(usedContUpdates) == 0 {
		return EncodingInfo{}, util.NewError(util.ErrNoSwsOrContsUsed)
	}

	return EncodingInfo{
		nodeIdToIndex:   *om.New[int64, int](),
		usedSwitchFTs:   usedSwitchFTs,
		usedContUpdates: usedContUpdates,
		links:           links,
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
			flowTables:  *om.New[int64, *convert.FlowTable](),
			frSequences: []*convert.FlowRuleSequence{},
		}
		for pair := c.NewFlowTables().Oldest(); pair != nil; pair = pair.Next() {
			cUpdate.flowTables.Set(pair.Key, pair.Value.Filter(ftPred))
		}
		for _, frs := range c.NewFlowRuleSequences() {
			cUpdate.frSequences = append(cUpdate.frSequences, frs.Filter(frsPred))
		}

		if cUpdate.flowTables.Len() > 0 || len(cUpdate.frSequences) > 0 {
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
	for _, update := range ei.usedContUpdates {
		newFt, exists := update.flowTables.Get(nodeId)
		if exists {
			return newFt, true
		}
	}

	return nil, false
}

func (ei EncodingInfo) GetSwIndex(nodeId int64) int {
	indx, exists := ei.nodeIdToIndex.Get(nodeId)
	if !exists {
		ei.nodeIdToIndex.Set(nodeId, ei.nodeIdToIndex.Len())
		indx, _ = ei.nodeIdToIndex.Get(nodeId)
	}

	return indx
}

func (ei EncodingInfo) GetSwIndexStr(nodeId int64) string {
	indx := ei.GetSwIndex(nodeId)
	return fmt.Sprintf("%d", indx)
}
