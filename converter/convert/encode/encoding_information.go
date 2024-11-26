package encode

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type EncodingInfo struct {
	nodeIdToIndex om.OrderedMap[int64, int]                  // maps switch node id to index
	usedSwitchFTs om.OrderedMap[int64, *convert.FlowTable]   // maps switch node id to flow table of switch
	usedContFTs   []om.OrderedMap[int64, *convert.FlowTable] // maps switch node id to new flow table
}

func NewEncodingInfo(n *convert.Network) (EncodingInfo, error) {
	usedSwitchFTs := getUsedSwitchesFTs(n.Switches())
	usedControllerFTs := getUsedControllers(n.Controllers())

	if usedSwitchFTs.Len() == 0 || len(usedControllerFTs) == 0 {
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
) []om.OrderedMap[int64, *convert.FlowTable] {
	usedControllerFTs := []om.OrderedMap[int64, *convert.FlowTable]{}

	for _, c := range controllers {
		if c.NewFlowTables().Len() > 0 {
			usedControllerFTs = append(usedControllerFTs, *c.NewFlowTables())
		}
	}

	return usedControllerFTs
}

func (ei EncodingInfo) FindNewFT(nodeId int64) (*convert.FlowTable, bool) {
	for _, newFTs := range ei.usedContFTs {
		newFt, exists := newFTs.Get(nodeId)
		if exists {
			return newFt, true
		}
	}

	return nil, false
}
