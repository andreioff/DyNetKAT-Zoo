package convert

import (
	"errors"
	"maps"
	"math/rand"
	"slices"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/path"
	"utwente.nl/topology-to-dynetkat-coverter/util"
)

const SEED int64 = 3

var randGen rand.Rand

func init() {
	randGen = *rand.New(rand.NewSource(SEED))
}

type Network struct {
	topology      util.Graph
	shortestPaths map[util.I64Tup][]*Switch // maps a tuple of topology node ids to a switch path

	switches []Switch
	hosts    []Host
	portNr   int64
	hostId   int64
}

func NewNetwork(topo util.Graph) (*Network, error) {
	var portNr int64 = 0

	switches, err := makeSwitchesFromTopology(topo, &portNr)
	if err != nil {
		return &Network{}, err
	}

	return &Network{
		topology:      topo,
		shortestPaths: computeShortestPaths(topo, switches),
		switches:      switches,
		portNr:        portNr,
		hostId:        0,
		hosts:         []Host{},
	}, nil
}

func (n *Network) PortNr() int64 {
	return n.portNr
}

func (n *Network) Switches() []Switch {
	return n.switches
}

func makeSwitchesFromTopology(topo util.Graph, portNr *int64) ([]Switch, error) {
	if portNr == nil {
		return []Switch{}, errors.New("Nil portNr argument!")
	}

	switches := []Switch{}

	iter := topo.Nodes()
	for iter.Next() {
		links, err := makeLinks(topo, iter.Node(), portNr)
		if err != nil {
			return []Switch{}, err
		}

		newSw := *NewSwitch(iter.Node(), links)
		switches = append(switches, newSw)
	}

	return switches, nil
}

func makeLinks(topo util.Graph, node graph.Node, portNr *int64) ([]Link, error) {
	if portNr == nil {
		return []Link{}, errors.New("Nil portNr argument!")
	}

	incidentEdges, err := util.GetIncidentEdges(topo, node)
	if err != nil {
		return []Link{}, err
	}

	links := []Link{}
	for _, edge := range incidentEdges {
		newLink := *NewLink(edge, *portNr, *portNr+1)
		links = append(links, newLink)
		*portNr += 2
	}
	return links, nil
}

func mapNodeToSwitch(switches []Switch) map[int64]*Switch {
	nodeIdToSwitch := make(map[int64]*Switch, len(switches))

	for _, sw := range switches {
		nodeIdToSwitch[sw.topoNode.ID()] = &sw
	}

	return nodeIdToSwitch
}

func nodePathToSwitchPath(path []graph.Node, nodeToSwitch map[int64]*Switch) []*Switch {
	switchPath := []*Switch{}
	for _, node := range path {
		switchPath = append(switchPath, nodeToSwitch[node.ID()])
	}
	return switchPath
}

func computeShortestPaths(topo util.Graph, switches []Switch) map[util.I64Tup][]*Switch {
	nodePaths := path.DijkstraAllPaths(&topo)
	nodeToSwitch := mapNodeToSwitch(switches)
	switchPaths := make(map[util.I64Tup][]*Switch)

	for i := range len(switches) {
		sw1Id := switches[i].topoNode.ID()

		for j := range len(switches) {
			if i == j {
				continue
			}

			sw2Id := switches[j].topoNode.ID()
			nodePath, _, _ := nodePaths.Between(sw1Id, sw2Id)
			switchPaths[util.NewI64Tup(sw1Id, sw2Id)] = nodePathToSwitchPath(
				nodePath,
				nodeToSwitch,
			)
		}
	}

	return switchPaths
}

func (n *Network) assignHosts(hostsNr int64) error {
	hosts := []Host{}

	randSws, err := n.pickRandomSwitches(hostsNr)
	if err != nil {
		return err
	}

	for _, randSw := range randSws {
		newHost, err := NewHost(n.hostId, n.portNr, randSw)
		if err != nil {
			return err
		}
		hosts = append(hosts, newHost)
		randSw.AddHost(&newHost)

		n.hostId++
		n.portNr++
	}
	n.hosts = hosts

	return nil
}

// Turns out that the switches order in the array is not static,
// so we must pick them by ID
func (n *Network) pickRandomSwitches(picksNr int64) ([]*Switch, error) {
	if len(n.switches) == 0 {
		return []*Switch{}, errors.New("Network has no switches!")
	}
	nodeToSw := mapNodeToSwitch(n.switches)
	nodeIds := slices.Collect(maps.Keys(nodeToSw))
	minId := slices.Min(nodeIds)
	maxId := int(slices.Max(nodeIds))

	if minId < 0 {
		panic("Found negative node ID when picking random switches!")
	}

	randSws := []*Switch{}
	for picksNr > 0 {
		randSwId := int64(randGen.Intn(maxId))

		sw, exists := nodeToSw[randSwId]
		if !exists {
			continue
		}
		randSws = append(randSws, sw)
		picksNr--
	}

	return randSws, nil
}

func (n *Network) populateDestinationTables(h1, h2 *Host) error {
	if h1 == nil || h2 == nil {
		return errors.New("Null arguments!")
	}

	path := n.shortestPaths[util.NewI64Tup(h1.sw.topoNode.ID(), h2.sw.topoNode.ID())]

	receivingPort := h1.SwitchPort()
	for i := range len(path) - 1 {
		currSw := path[i]
		nextSwId := path[i+1].topoNode.ID()
		link := currSw.FindLink(nextSwId)

		currSw.AddDestEntry(h2.ID(), receivingPort, link.FromPort())
		currSw.AddDestEntry(h2.ID(), link.FromPort(), link.ToPort())
		receivingPort = link.ToPort()
	}

	h2.sw.AddDestEntry(h2.ID(), receivingPort, h2.SwitchPort())
	return nil
}

func (n *Network) AddAndConnectHosts(hostsNr int64) error {
	if hostsNr < 2 {
		return errors.New("Number of hosts must be at least 2!")
	}

	err := n.assignHosts(hostsNr)
	if err != nil {
		return err
	}

	for i := range len(n.hosts) {
		for j := i + 1; j < len(n.hosts); j++ {
			err := n.populateDestinationTables(&n.hosts[i], &n.hosts[j])
			if err != nil {
				return err
			}

			err = n.populateDestinationTables(&n.hosts[j], &n.hosts[i])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
