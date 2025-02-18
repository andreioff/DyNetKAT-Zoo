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
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedContUpdates := getUsedControllers(n.Controllers())

	if usedSwitchFTs.Len() == 0 || len(usedContUpdates) == 0 {
		return EncodingInfo{}, util.NewError(util.ErrNoSwsOrContsUsed)
	}

	return EncodingInfo{
		nodeIdToIndex:   *om.New[int64, int](),
		usedSwitchFTs:   usedSwitchFTs,
		usedContUpdates: usedContUpdates,
	}, nil
}

func getUsedSwitchesFTs(switches []*convert.Switch) om.OrderedMap[int64, *convert.FlowTable] {
	usedSwitchFTs := *om.New[int64, *convert.FlowTable]()

	for _, sw := range switches {
		c := sw.Controller()
		if sw.FlowTable().Entries().Len() > 0 || c.IsUpdatingSwitch(sw.TopoNode().ID()) {
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
