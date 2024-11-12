package behavior

import (
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type Behavior interface {
	ModifyNetwork(n *convert.Network) error
}

func NewNetworkWithBehavior(topo util.Graph, b Behavior) (*convert.Network, error) {
	newNet, err := convert.NewNetwork(topo)
	if err != nil {
		return newNet, err
	}

	net := *newNet              // copy the value at the pointer's memory location
	err = b.ModifyNetwork(&net) // apply the modification on the copy
	if err != nil {
		// if something goes bad, return the initial, empty network
		return newNet, err
	}
	return &net, nil
}
