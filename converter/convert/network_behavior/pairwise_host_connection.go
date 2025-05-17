package behavior

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

type HostPair = util.Tuple[*convert.Host, *convert.Host]

func NewHostPair(h1 *convert.Host, h2 *convert.Host) HostPair {
	return HostPair{Fst: h1, Snd: h2}
}

/*
Given an empty network, assigns a set of flow rule sequences
that will be installed on the corresponding switches to
establish the connection of a given number of hosts with
each other. Each sequence contains the necessary flow rules to
establish the connection between a pair of hosts.
*/
type PairwiseHostConn struct {
	net                *convert.Network
	alternateDirection bool
}

/*
alternateDirection: whether to alternate the direction of the host connection between controllers.
By default, each controller is assigned a unique host pair and establishes
a one way connection between them, i.e. from host1 to host2.
*/
func NewPairwiseHostConn(alternateDirection bool) *PairwiseHostConn {
	return &PairwiseHostConn{net: nil, alternateDirection: alternateDirection}
}

func (phc *PairwiseHostConn) ModifyNetwork(n *convert.Network, config BehaviorConfig) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}
	phc.net = n

	err := n.AddControllersNoSplit(config.Controllers_nr)
	if err != nil {
		return err
	}

	newHostPairs, err := phc.createHostPairs()
	if err != nil {
		return err
	}
	return phc.populateControllerNewFTs(newHostPairs)
}

func (phc *PairwiseHostConn) createHostPairs() ([]HostPair, error) {
	swIdPair, err := phc.net.GetFarthestApartSwitches()
	if err != nil {
		return []HostPair{}, err
	}

	n := len(phc.net.Controllers())
	if phc.alternateDirection {
		n = (len(phc.net.Controllers()) + 1) / 2
	}

	hosts := []HostPair{}
	for range n {
		host1, err := phc.net.CreateHost(swIdPair.Fst)
		if err != nil {
			return []HostPair{}, err
		}

		host2, err := phc.net.CreateHost(swIdPair.Snd)
		if err != nil {
			return []HostPair{}, err
		}

		hosts = append(hosts, NewHostPair(host1, host2))
	}
	return hosts, nil
}

func (phc *PairwiseHostConn) populateControllerNewFTs(hostPairs []HostPair,
) error {
	hpIndex := 0
	for i, controller := range phc.net.Controllers() {
		host1, host2 := hostPairs[hpIndex].Fst, hostPairs[hpIndex].Snd
		if phc.alternateDirection && i%2 == 1 {
			err := phc.addHostPairConnFRs(controller, host2, host1)
			if err != nil {
				return err
			}
			hpIndex += 1
			continue
		}
		err := phc.addHostPairConnFRs(controller, host1, host2)
		if err != nil {
			return err
		}
		if !phc.alternateDirection {
			hpIndex += 1
		}
	}
	return nil
}

func (phc *PairwiseHostConn) addHostPairConnFRs(
	ct *convert.Controller, src, dest *convert.Host,
) error {
	switch {
	case src == nil:
		return util.NewError(util.ErrNilArgument, "src")
	case dest == nil:
		return util.NewError(util.ErrNilArgument, "dest")
	case ct == nil:
		return util.NewError(util.ErrNilArgument, "ct")
	}

	entries, err := phc.net.GetFlowRulesForSwitchPath(
		src.Switch(),
		dest.Switch(),
		src.SwitchPort(),
		dest.SwitchPort(),
	)
	if err != nil {
		return err
	}

	err = phc.addFRsToController(ct, dest.ID(), entries)
	if err != nil {
		return err
	}

	return nil
}

func (phc *PairwiseHostConn) addFRsToController(
	ct *convert.Controller,
	destHostId int64,
	entries om.OrderedMap[int64, []convert.FlowRule],
) error {
	for pair := entries.Oldest(); pair != nil; pair = pair.Next() {
		nodeId, frs := pair.Key, pair.Value
		err := ct.AddNewFlowRules(nodeId, destHostId, frs, false)
		if err != nil {
			return err
		}
	}

	return nil
}
