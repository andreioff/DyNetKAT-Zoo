package encode

import (
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
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedContUpdates := getUsedControllers(n.Controllers())

	if usedSwitchFTs.Len() == 0 || len(usedContUpdates) == 0 {
		return EncodingInfo{}, util.NewError(util.ErrNoSwsOrContsUsed)
	}

	return EncodingInfo{
		nodeIdToIndex:   getNodeIdToIndex(n.Switches(), usedSwitchFTs),
		usedSwitchFTs:   usedSwitchFTs,
		usedContUpdates: usedContUpdates,
	}, nil
}

func getNodeIdToIndex(
	switches []*convert.Switch,
	usedSwitchFTs om.OrderedMap[int64, *convert.FlowTable],
) om.OrderedMap[int64, int] {
	nodeIdToIndex := *om.New[int64, int]()
	index := 0
	for _, sw := range switches {
		_, exists := usedSwitchFTs.Get(sw.TopoNode().ID())
		if exists {
			nodeIdToIndex.Set(sw.TopoNode().ID(), index)
			index++
		}
	}
	return nodeIdToIndex
}

func getUsedSwitchesFTs(switches []*convert.Switch) om.OrderedMap[int64, *convert.FlowTable] {
	usedSwitchFTs := *om.New[int64, *convert.FlowTable]()

	for _, sw := range switches {
		c := sw.Controller()
		willReceiveUpdate := false
		if c != nil {
			_, willReceiveUpdate = c.NewFlowTables().Get(sw.TopoNode().ID())
		}

		if sw.FlowTable().Entries().Len() > 0 || willReceiveUpdate {
			usedSwitchFTs.Set(sw.TopoNode().ID(), sw.FlowTable())
		}
	}

	return usedSwitchFTs
}

func getUsedControllers(
	controllers []*convert.Controller,
) []ControllerUpdate {
	usedControllerUpdates := []ControllerUpdate{}

	for _, c := range controllers {
		if c.NewFlowTables().Len() > 0 || len(c.NewFlowRuleSequences()) > 0 {
			usedControllerUpdates = append(usedControllerUpdates, ControllerUpdate{
				flowTables:  *c.NewFlowTables(),
				frSequences: c.NewFlowRuleSequences(),
			})
		}
	}

	return usedControllerUpdates
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
