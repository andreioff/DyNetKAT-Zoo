package behavior

import (
	om "github.com/wk8/go-ordered-map/v2"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

/*
Given an empty network, assigns a set of flow rule sequences
that will be installed on the corresponding switches to
establish the connection of a given number of hosts with
each other. Each sequence contains the necessary flow rules to
establish the connection between a pair of hosts.
*/
type PairwiseHostConn struct {
	net *convert.Network
}

func (phc *PairwiseHostConn) ModifyNetwork(n *convert.Network, config BehaviorConfig) error {
	if n == nil {
		return util.NewError(util.ErrNilArgument, "n")
	}
	phc.net = n

	err := n.AddControllers(config.Controllers_nr)
	if err != nil {
		return err
	}

	newHosts, err := n.CreateRandomHosts(config.Outside_hosts_nr)
	if err != nil {
		return err
	}

	return phc.populateControllerNewFRSeqs(newHosts)
}

func (phc *PairwiseHostConn) populateControllerNewFRSeqs(
	newHosts []*convert.Host,
) error {
	for i, h1 := range newHosts {
		for j, h2 := range newHosts {
			if i >= j {
				continue
			}
			err := phc.addHostPairConnFRSeq(h1, h2)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (phc *PairwiseHostConn) addHostPairConnFRSeq(
	src, dest *convert.Host,
) error {
	switch {
	case src == nil:
		return util.NewError(util.ErrNilArgument, "src")
	case dest == nil:
		return util.NewError(util.ErrNilArgument, "dest")
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

	controllerSeqs, err := phc.makeFRSequences(dest.ID(), entries)
	if err != nil {
		return err
	}

	err = phc.addFRSequencesToControllers(controllerSeqs)
	if err != nil {
		return err
	}

	return nil
}

func (phc *PairwiseHostConn) makeFRSequences(
	destHostId int64,
	entries om.OrderedMap[int64, []convert.FlowRule],
) (om.OrderedMap[int64, *convert.FlowRuleSequence], error) {
	controllerSeqs := *om.New[int64, *convert.FlowRuleSequence]()

	for pair := entries.Oldest(); pair != nil; pair = pair.Next() {
		nodeId, frs := pair.Key, pair.Value
		sw, err := phc.net.GetSwitch(nodeId)
		if err != nil {
			return controllerSeqs, err
		}

		c := sw.Controller()
		if c == nil {
			return controllerSeqs, util.NewError(util.ErrSwitchHasNilController)
		}

		seq, exists := controllerSeqs.Get(c.ID())
		if !exists {
			seq = convert.NewFlowRuleSequence()
			controllerSeqs.Set(c.ID(), seq)
		}
		for _, fr := range frs {
			seq.AddEntry(nodeId, destHostId, fr)
		}
	}

	return controllerSeqs, nil
}

func (phc *PairwiseHostConn) addFRSequencesToControllers(
	controllerSeqs om.OrderedMap[int64, *convert.FlowRuleSequence],
) error {
	for pair := controllerSeqs.Oldest(); pair != nil; pair = pair.Next() {
		cid, seq := pair.Key, pair.Value
		c, err := phc.net.GetController(cid)
		if err != nil {
			return err
		}

		err = c.AddNewFlowRuleSequence(seq)
		if err != nil {
			return err
		}
	}
	return nil
}
