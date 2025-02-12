package behavior

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"utwente.nl/topology-to-dynetkat-coverter/convert"
	"utwente.nl/topology-to-dynetkat-coverter/util"
	ug "utwente.nl/topology-to-dynetkat-coverter/util/undirected_graph"
)

type TestFRSequence struct {
	sequence *convert.FlowRuleSequence
}

func newTestFRSequence() TestFRSequence {
	return TestFRSequence{convert.NewFlowRuleSequence()}
}

func (tfrs TestFRSequence) addTestEntry(
	switchId int64,
	destHostId int64,
	fr convert.FlowRule,
) TestFRSequence {
	tfrs.sequence.AddEntry(switchId, destHostId, fr)
	return tfrs
}

func assertEqualFRSequenceArrays(
	t *testing.T,
	expTestSeqs []TestFRSequence,
	actSeqs []*convert.FlowRuleSequence,
) {
	assert.Equal(t, len(expTestSeqs), len(actSeqs))
	for i, expTestSeq := range expTestSeqs {
		expSeq := expTestSeq.sequence
		assert.Equal(t, len(expSeq.Entries()), len(actSeqs[i].Entries()))
		for j, expEntry := range expSeq.Entries() {
			actEntry := actSeqs[i].Entries()[j]
			assert.EqualValues(t, expEntry, actEntry)
		}
	}
}

func TestBehavior_PairwiseHostConnection(t *testing.T) {
	cases := map[string]struct {
		randomSeed        int64
		topology          ug.WeightedUndirectedGraph
		config            BehaviorConfig
		expectedSequences map[int64][]TestFRSequence
	}{
		"2 Outside Hosts, 1 Controller [Success]": {
			randomSeed: 4, // connects host0 to switch 1, and host1 to switch 4
			topology:   getMockTopology1(),
			config: BehaviorConfig{
				Hosts_nr:         0,
				Outside_hosts_nr: 2,
				Controllers_nr:   1,
			},
			expectedSequences: map[int64][]TestFRSequence{
				0: {
					newTestFRSequence().
						addTestEntry(4, 1, convert.NewFlowRule(12, 7, false)).
						addTestEntry(4, 1, convert.NewFlowRule(7, 6, true)).
						addTestEntry(2, 1, convert.NewFlowRule(6, 3, false)).
						addTestEntry(2, 1, convert.NewFlowRule(3, 2, true)).
						addTestEntry(0, 1, convert.NewFlowRule(2, 0, false)).
						addTestEntry(0, 1, convert.NewFlowRule(0, 1, true)).
						addTestEntry(1, 1, convert.NewFlowRule(1, 13, false)),

					newTestFRSequence().
						addTestEntry(1, 0, convert.NewFlowRule(13, 1, false)).
						addTestEntry(1, 0, convert.NewFlowRule(1, 0, true)).
						addTestEntry(0, 0, convert.NewFlowRule(0, 2, false)).
						addTestEntry(0, 0, convert.NewFlowRule(2, 3, true)).
						addTestEntry(2, 0, convert.NewFlowRule(3, 6, false)).
						addTestEntry(2, 0, convert.NewFlowRule(6, 7, true)).
						addTestEntry(4, 0, convert.NewFlowRule(7, 12, false)),
				},
			},
		},
		"3 Outside Hosts, 2 Controllers [Success]": {
			// connects host0 to switch 5, and host1 to switch 2, and host2 to switch 1
			randomSeed: 15,
			topology:   getMockTopology1(),
			config: BehaviorConfig{
				Hosts_nr:         0,
				Outside_hosts_nr: 3,
				Controllers_nr:   2,
			},
			/*
				6 sequences, each split between 2 controllers because
				they control different switches in the network:
				  Controller 0 is responsible for switches 2, 3 and 4
				  Controller 1 is responsible for switches 0, 1 and 5.
			*/
			expectedSequences: map[int64][]TestFRSequence{
				0: {
					newTestFRSequence(). // 2nd part of seq 1: h0 to h1
								addTestEntry(4, 1, convert.NewFlowRule(10, 7, false)).
								addTestEntry(4, 1, convert.NewFlowRule(7, 6, true)).
								addTestEntry(2, 1, convert.NewFlowRule(6, 13, false)),

					newTestFRSequence(). // 2nd part of seq 2: h0 to h2
								addTestEntry(3, 2, convert.NewFlowRule(8, 5, false)).
								addTestEntry(3, 2, convert.NewFlowRule(5, 4, true)),

					newTestFRSequence(). // 1st part of seq 3: h1 to h0
								addTestEntry(2, 0, convert.NewFlowRule(13, 6, false)).
								addTestEntry(2, 0, convert.NewFlowRule(6, 7, true)).
								addTestEntry(4, 0, convert.NewFlowRule(7, 10, false)).
								addTestEntry(4, 0, convert.NewFlowRule(10, 11, true)),

					newTestFRSequence(). // 1st part of seq 4: h1 to h2
								addTestEntry(2, 2, convert.NewFlowRule(13, 3, false)).
								addTestEntry(2, 2, convert.NewFlowRule(3, 2, true)),

					newTestFRSequence(). // 2nd part of seq 5: h2 to h0
								addTestEntry(3, 0, convert.NewFlowRule(5, 8, false)).
								addTestEntry(3, 0, convert.NewFlowRule(8, 9, true)),

					newTestFRSequence(). // 2nd part of seq 6: h2 to h1
								addTestEntry(2, 1, convert.NewFlowRule(3, 13, false)),
				},
				1: {
					newTestFRSequence(). // 1st part of seq 1: h0 to h1
								addTestEntry(5, 1, convert.NewFlowRule(12, 11, false)).
								addTestEntry(5, 1, convert.NewFlowRule(11, 10, true)),

					newTestFRSequence(). // 1st and 3rd part of seq 2: h0 to h2
								addTestEntry(5, 2, convert.NewFlowRule(12, 9, false)).
								addTestEntry(5, 2, convert.NewFlowRule(9, 8, true)).
								addTestEntry(1, 2, convert.NewFlowRule(4, 14, false)),

					newTestFRSequence(). // 2nd part of seq 3: h1 to h0
								addTestEntry(5, 0, convert.NewFlowRule(11, 12, false)),

					newTestFRSequence(). // 2st part of seq 4: h1 to h2
								addTestEntry(0, 2, convert.NewFlowRule(2, 0, false)).
								addTestEntry(0, 2, convert.NewFlowRule(0, 1, true)).
								addTestEntry(1, 2, convert.NewFlowRule(1, 14, false)),

					newTestFRSequence(). // 1st and 3rd part of seq 5: h2 to h0
								addTestEntry(1, 0, convert.NewFlowRule(14, 4, false)).
								addTestEntry(1, 0, convert.NewFlowRule(4, 5, true)).
								addTestEntry(5, 0, convert.NewFlowRule(9, 12, false)),

					newTestFRSequence(). // 1st part of seq 6: h2 to h1
								addTestEntry(1, 1, convert.NewFlowRule(14, 1, false)).
								addTestEntry(1, 1, convert.NewFlowRule(1, 0, true)).
								addTestEntry(0, 1, convert.NewFlowRule(0, 2, false)).
								addTestEntry(0, 1, convert.NewFlowRule(2, 3, true)),
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			util.SetRandGenSeed(tc.randomSeed)

			n, err := NewNetworkWithBehavior(getMockTopology1(), &PairwiseHostConn{}, tc.config)
			assert.NoError(t, err)

			// all hosts must be outside hosts, i.e. not connected to the network
			assert.Equal(t, 0, len(n.Hosts()))
			for _, s := range n.Switches() {
				assert.Equal(t, 0, s.FlowTable().Entries().Len())
			}
			for _, c := range n.Controllers() {
				assert.Equal(t, 0, c.NewFlowTables().Len())
				expSeqs, exists := tc.expectedSequences[c.ID()]
				if !exists {
					assert.Equal(t, 0, len(c.NewFlowRuleSequences()))
					continue
				}
				assertEqualFRSequenceArrays(t, expSeqs, c.NewFlowRuleSequences())
			}
		})
	}
}
