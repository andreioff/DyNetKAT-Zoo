package encode

import (
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type EncodingInfo struct {
	nodeIdToIndex map[int64]int                  // maps switch node id to index
	usedSwitchFTs map[int64]*convert.FlowTable   // maps switch node id to flow table of switch
	usedContFTs   []map[int64]*convert.FlowTable // maps switch node id to new flow table
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedControllerFTs := getUsedControllers(n.Controllers())

	if len(usedSwitchFTs) == 0 || len(usedControllerFTs) == 0 {
		return EncodingInfo{}, util.NewError(util.ErrNoSwsOrContsUsed)
	}

	return EncodingInfo{
		nodeIdToIndex: getNodeIdToIndex(n.Switches(), usedSwitchFTs),
		usedSwitchFTs: usedSwitchFTs,
		usedContFTs:   usedControllerFTs,
	}, nil
}

func getNodeIdToIndex(
	switches []*convert.Switch,
	usedSwitchFTs map[int64]*convert.FlowTable,
) map[int64]int {
	nodeIdToIndex := make(map[int64]int)
	index := 0
	for _, sw := range switches {
		_, exists := usedSwitchFTs[sw.TopoNode().ID()]
		if exists {
			nodeIdToIndex[sw.TopoNode().ID()] = index
			index++
		}
	}
	return nodeIdToIndex
}

func getUsedSwitchesFTs(switches []*convert.Switch) map[int64]*convert.FlowTable {
	usedSwitchFTs := make(map[int64]*convert.FlowTable)

	for _, sw := range switches {
		c := sw.Controller()
		willReceiveUpdate := false
		if c != nil {
			_, willReceiveUpdate = c.NewFlowTables()[sw.TopoNode().ID()]
		}

		if len(sw.FlowTable().Entries()) > 0 || willReceiveUpdate {
			usedSwitchFTs[sw.TopoNode().ID()] = sw.FlowTable()
		}
	}

	return usedSwitchFTs
}

func getUsedControllers(controllers []*convert.Controller) []map[int64]*convert.FlowTable {
	usedControllerFTs := []map[int64]*convert.FlowTable{}

	for _, c := range controllers {
		if len(c.NewFlowTables()) > 0 {
			usedControllerFTs = append(usedControllerFTs, c.NewFlowTables())
		}
	}

	return usedControllerFTs
}

func (ei EncodingInfo) FindNewFT(nodeId int64) (*convert.FlowTable, bool) {
	for _, newFTs := range ei.usedContFTs {
		newFt, exists := newFTs[nodeId]
		if exists {
			return newFt, true
		}
	}

	return nil, false
}
